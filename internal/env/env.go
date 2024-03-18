package env

import (
	"errors"

	"github.com/caarlos0/env/v10"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	TimeZone          string  `env:"TZ" envDefault:"Etc/UTC" validate:"timezone"`
	PublicUrl         string  `env:"PUBLIC_URL" validate:"url"`
	Title             string  `env:"TITLE" envDefault:"goDash"`
	Port              int     `env:"PORT" envDefault:"4000" validate:"min=1024,max=49151"`
	Version           string  `env:"APP_VERSION"`
	LocationLatitude  float32 `env:"LOCATION_LATITUDE" envDefault:"48.780331609463815"`
	LocationLongitude float32 `env:"LOCATION_LONGITUDE" envDefault:"9.177968320179422"`
	WeatherKey        string  `env:"WEATHER_KEY"`
	WeatherUnits      string  `env:"WEATHER_UNITS" envDefault:"metric"`
	WeatherLanguage   string  `env:"WEATHER_LANG" envDefault:"en"`
	WeatherDigits     bool    `env:"WEATHER_DIGITS" envDefault:"true"`
	LogtoEndpoint     string  `env:"LOGTO_ENDPOINT" validate:"url"`
	LogtoAppId        string  `env:"LOGTO_APP_ID,unset"`
	LogtoAppSecret    string  `env:"LOGTO_APP_SECRET,unset"`
	SessionKey        string  `env:"SESSION_KEY,unset"`
}

var errParse = errors.New("error parsing environment variables")

func Parse() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	if err := validateContent(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func validateContent(cfg *Config) error {
	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		} else {
			for _, err := range err.(validator.ValidationErrors) {
				return err
			}
		}
		return errParse
	}
	return nil
}
