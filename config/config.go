package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flohoss/godash/internal/validate"
	"github.com/flohoss/godash/pkg/media"
	"github.com/flohoss/gosso/v2"
	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const (
	ConfigFolder = "./config/"
	iconsFolder  = ConfigFolder + "icons/"
)

var (
	iconCache   = map[string][2]string{}
	iconCacheMu sync.Mutex
)

var cfg GlobalConfig
var mu sync.RWMutex
var reloadMu sync.Mutex

type GlobalConfig struct {
	LogLevel     string         `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error"`
	TimeZone     string         `mapstructure:"time_zone" validate:"omitempty,timezone"`
	Title        string         `mapstructure:"title"`
	Server       ServerSettings `mapstructure:"server"`
	Weather      Weather        `mapstructure:"weather"`
	Applications []Category     `mapstructure:"applications"`
	Links        []Category     `mapstructure:"links"`
	Auth         AuthSettings   `mapstructure:"auth" validate:"omitempty"`
}

const MinSessionKeyLength = gosso.MinSessionKeyLength

// validateSessionKey re-exports gosso's session-key rule so callers configuring
// godash directly keep the same enforcement on the app level.
func validateSessionKey(key string, enabled bool) error {
	if !enabled {
		return nil
	}

	if err := gosso.ValidateSessionKey(key); err != nil {
		return fmt.Errorf("configuration validation failed: auth.oidc.session_key must be at least %d characters when OIDC is enabled: %w", MinSessionKeyLength, err)
	}

	return nil
}

const (
	ProviderOIDC = "oidc"

	oidcProviderName = "OIDC"
	oidcProviderIcon = "icon-[cib--open-id]"
)

type AuthSettings struct {
	OIDC OIDCSettings `mapstructure:"oidc" validate:"omitempty"`
}

type AuthOption struct {
	Provider string
	Name     string
	Icon     string
}

func (o AuthOption) Path() string {
	return "/auth/" + o.Provider
}

type OIDCSettings struct {
	Enabled      bool          `mapstructure:"enabled"`
	IssuerURL    string        `mapstructure:"issuer_url" validate:"required_if=Enabled true,omitempty,url,endsnotwith=/"`
	ClientID     string        `mapstructure:"client_id" validate:"required_if=Enabled true"`
	ClientSecret string        `mapstructure:"client_secret" validate:"required_if=Enabled true"`
	SessionKey   string        `mapstructure:"session_key" validate:"omitempty"`
	SessionTTL   time.Duration `mapstructure:"session_ttl" validate:"gte=0"`
	CookieSecure bool          `mapstructure:"cookie_secure"`
}

type ServerSettings struct {
	Address        string   `mapstructure:"address" validate:"omitempty,ipv4"`
	Port           int      `mapstructure:"port" validate:"omitempty,gte=1024,lte=65535"`
	TrustedProxies []string `mapstructure:"trusted_proxies" validate:"omitempty,dive,cidr"`
	AppUrl         string   `mapstructure:"app_url" validate:"omitempty,url,endsnotwith=/"`
}

type Weather struct {
	Units     string  `mapstructure:"units" validate:"omitempty,oneof=celsius fahrenheit"`
	Latitude  float64 `mapstructure:"latitude" validate:"omitempty,latitude"`
	Longitude float64 `mapstructure:"longitude" validate:"omitempty,longitude"`
}

type Category struct {
	Category string `mapstructure:"category"`
	Entries  []App  `mapstructure:"entries"`
}

type App struct {
	Name       string `mapstructure:"name"`
	Icon       string `mapstructure:"icon"`
	IconLight  string `mapstructure:"-"`
	URL        string `mapstructure:"url" validate:"omitempty,url"`
	IgnoreDark bool   `mapstructure:"ignore_dark"`
}

type AppConfig struct {
	Name       string `mapstructure:"name"`
	Icon       string `mapstructure:"icon"`
	URL        string `mapstructure:"url"`
	IgnoreDark bool   `mapstructure:"ignore_dark"`
}

func init() {
	os.Mkdir(ConfigFolder, 0750)
	os.Mkdir(iconsFolder, 0750)
}

func New() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("time_zone", "Europe/Berlin")
	viper.SetDefault("server.address", "0.0.0.0")
	viper.SetDefault("server.port", 8156)
	viper.SetDefault("title", "GoDash")
	viper.SetDefault("weather.units", "celsius")
	viper.SetDefault("weather.latitude", 52.5163)
	viper.SetDefault("weather.longitude", 13.3776)
	viper.SetDefault("auth.oidc.session_ttl", "24h")
	viper.SetDefault("applications", []map[string]any{
		{
			"category": "Applications",
			"entries": []map[string]any{
				{
					"name": "GoDash",
					"icon": "sh/homebox",
					"url":  "https://github.com/flohoss/godash",
				},
			},
		},
	})
	viper.SetDefault("links", []map[string]any{
		{
			"category": "Applications",
			"entries": []map[string]any{
				{
					"name": "GoDash",
					"url":  "https://github.com/flohoss/godash",
				},
			},
		},
	})

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(ConfigFolder)
	viper.SetEnvPrefix("GD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.WriteConfigAs(ConfigFolder + "config.yaml")
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
		} else {
			slog.Error("Failed to read configuration file", "error", err)
			os.Exit(1)
		}
	}

	if err := ValidateAndLoadConfig(); err != nil {
		slog.Error("Initial configuration validation failed", "error", err)
		os.Exit(1)
	}
}

func ValidateAndLoadConfig() error {
	reloadMu.Lock()
	defer reloadMu.Unlock()

	var tempCfg GlobalConfig
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
	)
	if err := viper.Unmarshal(&tempCfg, viper.DecodeHook(decodeHook)); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := validateSessionKey(tempCfg.Auth.OIDC.SessionKey, tempCfg.Auth.OIDC.Enabled); err != nil {
		return err
	}

	if err := validate.Struct(tempCfg); err != nil {
		return fmt.Errorf("configuration validation failed:\n%s", err)
	}

	clearIconCache()

	replaceIconStrings(tempCfg.Applications)
	replaceIconStrings(tempCfg.Links)

	cleanupOrphanedIcons(iconsFolder, tempCfg.Applications, tempCfg.Links)

	mu.Lock()
	cfg = tempCfg
	mu.Unlock()

	if tempCfg.Auth.OIDC.Enabled && tempCfg.Server.AppUrl == "" {
		slog.Warn("OIDC is enabled without server.app_url; the callback URL is derived from the request")
	}

	return nil
}

func replaceIconStrings(applications []Category) {
	for i := range applications {
		for j := range applications[i].Entries {
			bookmark := &applications[i].Entries[j]

			if cached, ok := lookupIconCache(bookmark.Icon); ok {
				bookmark.Icon = cached[0]
				bookmark.IconLight = cached[1]
				continue
			}

			var filePath, filePathLight string

			switch {
			case isExternalIcon(bookmark.Icon):
				if u, err := url.Parse(bookmark.Icon); err == nil && u.Scheme != "" && u.Host != "" {
					filePath, filePathLight = bookmark.Icon, ""
				}
			case strings.HasPrefix(bookmark.Icon, "sh/"):
				path, lightPath, err := downloadSelfHostedIcon(bookmark.Icon)
				if err != nil {
					slog.Error(err.Error())
				} else {
					filePath, filePathLight = path, lightPath
				}
			default:
				ext := filepath.Ext(bookmark.Icon)
				path, lightPath, err := handleLocalIcons(iconsFolder, bookmark.Icon, ext)
				if err != nil {
					slog.Warn("could not find local icon", "path", bookmark.Icon, "error", err)
				}
				filePath, filePathLight = path, lightPath
			}

			storeIconCache(bookmark.Icon, [2]string{filePath, filePathLight})
			bookmark.Icon = filePath
			bookmark.IconLight = filePathLight
		}
	}
}

func lookupIconCache(key string) ([2]string, bool) {
	iconCacheMu.Lock()
	defer iconCacheMu.Unlock()
	v, ok := iconCache[key]
	return v, ok
}

func storeIconCache(key string, value [2]string) {
	iconCacheMu.Lock()
	defer iconCacheMu.Unlock()
	iconCache[key] = value
}

func clearIconCache() {
	iconCacheMu.Lock()
	defer iconCacheMu.Unlock()
	iconCache = map[string][2]string{}
}

func cleanupOrphanedIcons(iconsDir string, categories ...[]Category) {
	referenced := map[string]bool{}
	for _, apps := range categories {
		for i := range apps {
			for _, app := range apps[i].Entries {
				for _, path := range []string{app.Icon, app.IconLight} {
					if path == "" || isExternalIcon(path) {
						continue
					}
					referenced[path] = true
				}
			}
		}
	}

	entries, err := os.ReadDir(iconsDir)
	if err != nil {
		slog.Warn("could not scan icons folder for cleanup", "error", err)
		return
	}
	for _, entry := range entries {
		path := "/icons/" + entry.Name()
		if !referenced[path] {
			if err := os.Remove(iconsDir + entry.Name()); err != nil {
				slog.Warn("could not remove orphaned icon", "path", entry.Name(), "error", err)
			} else {
				slog.Info("removed orphaned icon", "path", entry.Name())
			}
		}
	}
}

func isExternalIcon(icon string) bool {
	return strings.HasPrefix(icon, "http://") || strings.HasPrefix(icon, "https://")
}

func downloadSelfHostedIcon(icon string) (string, string, error) {
	for _, ext := range []string{".svg", ".webp", ".png"} {
		title, iconURL, lightTitle, lightURL := handleSelfHostedIcons(icon, ext)
		if title == "" {
			return "", "", fmt.Errorf("invalid self-hosted icon: %s", icon)
		}
		path, lightPath, err := downloadIcons(title, iconURL, lightTitle, lightURL)
		if err == nil {
			return path, lightPath, nil
		}
		slog.Debug("self-hosted icon format unavailable, trying next", "icon", icon, "ext", ext, "error", err)
	}
	return "", "", fmt.Errorf("no available format for self-hosted icon: %s", icon)
}

func downloadIcons(title, url, lightTitle, lightUrl string) (string, string, error) {
	path, err := downloadIcon(title, url)
	if err != nil {
		return "", "", err
	}
	lightPath, err := downloadIcon(lightTitle, lightUrl)
	if err != nil {
		slog.Warn("could not download light icon, falling back to base", "title", lightTitle, "error", err)
		lightPath = path
	}
	return path, lightPath, nil
}

func downloadIcon(title, url string) (string, error) {
	filePath := iconsFolder + title
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		filePath, err = media.DownloadSelfHostedIcon(url, title, filePath)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", fmt.Errorf("stat icon %q: %w", filePath, err)
	}
	return "/" + strings.TrimPrefix(filePath, ConfigFolder), nil
}

func handleSelfHostedIcons(icon, ext string) (string, string, string, string) {
	name := strings.Replace(icon, "sh/", "", 1)
	clean := filepath.Clean(name)
	if strings.Contains(clean, "..") || filepath.IsAbs(clean) {
		return "", "", "", ""
	}
	title := clean + ext
	encoded := url.PathEscape(title)
	iconURL := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + encoded
	lightTitle := strings.Replace(title, ext, "-light"+ext, 1)
	lightEncoded := url.PathEscape(lightTitle)
	lightURL := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + lightEncoded
	return title, iconURL, lightTitle, lightURL
}

func handleLocalIcons(iconsDir, title, ext string) (string, string, error) {
	clean := filepath.Clean(title)
	if strings.Contains(clean, "..") || filepath.IsAbs(clean) {
		return "", "", fmt.Errorf("invalid icon path: %s", title)
	}
	filePath := iconsDir + clean
	if _, err := os.Stat(filePath); err != nil {
		return "", "", err
	}
	basePath := "/" + strings.TrimPrefix(filePath, ConfigFolder)
	filePathLight := strings.Replace(filePath, ext, "-light"+ext, 1)
	if _, err := os.Stat(filePathLight); err != nil {
		return basePath, basePath, nil
	}
	return basePath, "/" + strings.TrimPrefix(filePathLight, ConfigFolder), nil
}

func ConfigLoaded() bool {
	return viper.ConfigFileUsed() != ""
}

func GetLogLevel() slog.Level {
	mu.RLock()
	defer mu.RUnlock()
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func GetServer() string {
	mu.RLock()
	defer mu.RUnlock()
	return net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port))
}

func GetApplications() []Category {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Applications
}

func GetLinks() []Category {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Links
}

func GetTitle() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Title
}

func GetTimeZone() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.TimeZone
}

func GetWeatherConfig() (Weather, string) {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Weather, cfg.TimeZone
}

func GetAuth() AuthSettings {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Auth
}

func GetAuthProviders() []AuthOption {
	mu.RLock()
	defer mu.RUnlock()

	providers := make([]AuthOption, 0, 1)
	if cfg.Auth.OIDC.Enabled {
		providers = append(providers, AuthOption{Provider: ProviderOIDC, Name: oidcProviderName, Icon: oidcProviderIcon})
	}

	return providers
}

func GetTrustedProxies() []string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Server.TrustedProxies
}

func GetAppUrl() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Server.AppUrl
}
