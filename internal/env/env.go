package env

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v8"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	TimeZone   string `env:"TZ" envDefault:"Etc/UTC" validate:"timezone"`
	Title      string `env:"TITLE" envDefault:"goDash"`
	Port       int    `env:"PORT" envDefault:"4000" validate:"min=1024,max=49151"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info" validate:"oneof=debug info warn error panic fatal"`
	LiveSystem bool   `env:"LIVE_SYSTEM" envDefault:"true"`
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
	setAllDefaultEnvs(cfg)
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

func setAllDefaultEnvs(cfg *Config) {
	os.Setenv("TZ", cfg.TimeZone)
	os.Setenv("PORT", fmt.Sprintf("%d", cfg.Port))
	os.Setenv("LOG_LEVEL", cfg.LogLevel)
}
