package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	sseserver "github.com/r3labs/sse/v2"
	"github.com/spf13/viper"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/handlers"
	"github.com/flohoss/godash/services"
)

func setupRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/sse"
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

	sse := sseserver.New()
	sse.AutoReplay = false

	s := services.NewSystemService(sse)
	w := services.NewWeatherService(sse)

	appHandler := handlers.NewAppHandler(s, w)
	handlers.SetupRoutes(e, sse, appHandler)

	slog.Info("Starting server", "url", fmt.Sprintf("http://%s", config.GetServer()))

	go func() {
		if err := e.Start(config.GetServer()); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server", "error", err)
	}
}
