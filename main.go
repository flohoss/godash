package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r3labs/sse/v2"
	"github.com/spf13/viper"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/handlers"
	"github.com/flohoss/godash/services"
	"github.com/flohoss/godash/views/system"
	"github.com/flohoss/godash/views/weather"
)

func setupRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "sse")
		},
	}))

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

	sse := sse.New()
	sse.AutoReplay = false

	s := services.NewSystemService(sse, func(id, icon, static string, detail services.Detail) templ.Component {
		return system.Badge(id, icon, static, detail)
	})
	w := services.NewWeatherService(sse,
		func(today services.Day) templ.Component {
			return weather.Current(today)
		},
		func(days []services.Day) templ.Component {
			return weather.Forecast(days)
		},
	)

	appHandler := handlers.NewAppHandler(s, w)
	handlers.SetupRoutes(e, sse, appHandler)

	slog.Info("Starting server", "url", fmt.Sprintf("http://%s", config.GetServer()))
	slog.Error(e.Start(config.GetServer()).Error())
}
