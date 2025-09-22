package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"gitlab.unjx.de/flohoss/godash/pkg/media"
)

const (
	ConfigFolder = "./config/"
	iconsFolder  = ConfigFolder + "icons/"
)

var cfg GlobalConfig

var validate *validator.Validate
var mu sync.RWMutex

type GlobalConfig struct {
	LogLevel     string         `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error"`
	TimeZone     string         `mapstructure:"time_zone" validate:"omitempty,timezone"`
	Title        string         `mapstructure:"title"`
	Server       ServerSettings `mapstructure:"server"`
	Weather      Weather        `mapstructure:"weather"`
	Applications []Category     `mapstructure:"applications"`
	Links        []Category     `mapstructure:"links"`
}

type ServerSettings struct {
	Address string `mapstructure:"address" validate:"omitempty,ipv4"`
	Port    int    `mapstructure:"port" validate:"omitempty,gte=1024,lte=65535"`
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
	os.Mkdir(ConfigFolder, os.ModePerm)
	os.Mkdir(iconsFolder, os.ModePerm)
	validate = validator.New()
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
	viper.SetDefault("applications", []map[string]interface{}{
		{
			"category": "Applications",
			"entries": []map[string]interface{}{
				{
					"name": "GoDash",
					"icon": "sh/homebox",
					"url":  "https://gitlab.unjx.de/flohoss/godash",
				},
			},
		},
	})
	viper.SetDefault("links", []map[string]interface{}{
		{
			"category": "Applications",
			"entries": []map[string]interface{}{
				{
					"name": "GoDash",
					"url":  "https://gitlab.unjx.de/flohoss/godash",
				},
			},
		},
	})

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(ConfigFolder)

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
	var tempCfg GlobalConfig
	if err := viper.Unmarshal(&tempCfg); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := validate.Struct(tempCfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	replaceIconStrings(tempCfg.Applications)
	replaceIconStrings(tempCfg.Links)

	mu.Lock()
	cfg = tempCfg
	mu.Unlock()

	os.Setenv("TZ", cfg.TimeZone)
	return nil
}

func replaceIconStrings(applications []Category) {
	for i := range applications {
		for j := range applications[i].Entries {
			bookmark := &applications[i].Entries[j]

			var filePath, filePathLight string
			var err error

			if strings.HasPrefix(bookmark.Icon, "sh/") {
				filePath, filePathLight, err = downloadIcons(handleSelfHostedIcons(bookmark.Icon, ".webp"))
				if err != nil {
					slog.Error(err.Error())
					continue
				}
			} else {
				ext := filepath.Ext(bookmark.Icon)
				filePath, filePathLight = handleLocalIcons(bookmark.Icon, ext)
				if filePath == "" {
					slog.Warn("could not find local icon", "path", bookmark.Icon)
				}
			}

			bookmark.Icon = filePath
			bookmark.IconLight = filePathLight
		}
	}
}

func downloadIcons(title, url, lightTitle, lightUrl string) (string, string, error) {
	path, err := downloadIcon(title, url)
	if err != nil {
		return "", "", err
	}
	lightPath, _ := downloadIcon(lightTitle, lightUrl)
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
	}
	return "/" + strings.TrimPrefix(filePath, ConfigFolder), nil
}

func handleSelfHostedIcons(icon, ext string) (string, string, string, string) {
	title := strings.Replace(icon, "sh/", "", 1) + ext
	url := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + title
	lightTitle := strings.Replace(title, ext, "-light"+ext, 1)
	lightUrl := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + lightTitle
	return title, url, lightTitle, lightUrl
}

func handleLocalIcons(title, ext string) (string, string) {
	filePath := iconsFolder + title
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", ""
	}
	filePathLight := strings.Replace(filePath, ext, "-light"+ext, 1)
	_, err = os.Stat(filePathLight)
	if os.IsNotExist(err) {
		return "/" + strings.TrimPrefix(filePath, ConfigFolder), ""
	}
	return "/" + strings.TrimPrefix(filePath, ConfigFolder), "/" + strings.TrimPrefix(filePathLight, ConfigFolder)
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
	return fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
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

func GetWeatherSettings() Weather {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Weather
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
