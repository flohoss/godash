package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	ConfigFolder = "./config/"
)

var cfg GlobalConfig

var validate *validator.Validate
var mu sync.RWMutex

type GlobalConfig struct {
	LogLevel     string         `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error"`
	TimeZone     string         `mapstructure:"time_zone" validate:"omitempty,timezone"`
	Title        string         `mapstructure:"title"`
	Server       ServerSettings `mapstructure:"server"`
	Location     Location       `mapstructure:"location"`
	Weather      Weather        `mapstructure:"weather"`
	Applications []Category     `mapstructure:"applications"`
}

type ServerSettings struct {
	Address string `mapstructure:"address" validate:"omitempty,ipv4"`
	Port    int    `mapstructure:"port" validate:"omitempty,gte=1024,lte=65535"`
}

type Location struct {
	Latitude  float32 `mapstructure:"latitude" validate:"omitempty,latitude"`
	Longitude float32 `mapstructure:"longitude" validate:"omitempty,longitude"`
}

type Weather struct {
	Key      string `mapstructure:"key"`
	Units    string `mapstructure:"units" validate:"omitempty,oneof=metric imperial"`
	Language string `mapstructure:"language" validate:"omitempty,bcp47_language_tag"`
}

type Category struct {
	Category string `mapstructure:"category"`
	Entries  []App  `mapstructure:"entries"`
}

type App struct {
	Name       string `mapstructure:"name"`
	Icon       string `mapstructure:"icon"`
	URL        string `mapstructure:"url" validate:"omitempty,url"`
	IgnoreDark bool   `mapstructure:"ignore_dark"`
}

func init() {
	os.Mkdir(ConfigFolder, os.ModePerm)
	validate = validator.New()
}

func New() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("time_zone", "Etc/UTC")
	viper.SetDefault("server.address", "0.0.0.0")
	viper.SetDefault("server.port", 8156)
	viper.SetDefault("title", "goDash")
	viper.SetDefault("weather.units", "metric")
	viper.SetDefault("weather.language", "en")
	viper.SetDefault("applications", []Category{})

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

	mu.Lock()
	cfg = tempCfg
	mu.Unlock()

	os.Setenv("TZ", cfg.TimeZone)
	return nil
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
