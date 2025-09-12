package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"

	"gitlab.unjx.de/flohoss/godash/config"
)

func setupRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	return e
}

func setLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.GetLogLevel(),
	}))
	slog.SetDefault(logger)
	slog.Debug("logger set", "level", config.GetLogLevel())
}

func setupViperWatcher() {
	var (
		mu    sync.Mutex
		timer *time.Timer
	)

	debounce := func(d time.Duration, fn func()) {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(d, fn)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		debounce(2*time.Second, func() {
			config.ValidateAndLoadConfig()
			setLogger()
			slog.Debug("config changed", "file", e.Name)
		})
	})

	viper.WatchConfig()
}

func main() {
	config.New()
	setLogger()

	e := setupRouter()

	setupViperWatcher()

	slog.Info("Starting server", "url", fmt.Sprintf("http://%s", config.GetServer()))
	slog.Error(e.Start(config.GetServer()).Error())
}
