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
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	sseserver "github.com/r3labs/sse/v2"
	"github.com/spf13/viper"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/handlers"
	"github.com/flohoss/godash/services"
)

func setupRouter(logger *slog.Logger) *echo.Echo {
	e := echo.NewWithConfig(echo.Config{
		Logger:      logger,
		IPExtractor: echo.ExtractIPFromRealIPHeader(),
	})

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c *echo.Context) bool {
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

func setupViperWatcher(echoInst *echo.Echo) {
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
			echoInst.Logger = slog.Default()
			slog.Debug("config changed", "file", e.Name)
		})
	})

	viper.WatchConfig()
}

func main() {
	config.New()
	setLogger()

	e := setupRouter(slog.Default())

	setupViperWatcher(e)

	sse := sseserver.New()
	sse.AutoReplay = false
	sse.OnSubscribe = func(streamID string, sub *sseserver.Subscriber) {
		services.PublishSnapshot(streamID)
	}

	s := services.NewSystemService(sse)
	w := services.NewWeatherService(sse)

	appHandler := handlers.NewAppHandler(s, w)
	handlers.SetupRoutes(e, sse, appHandler)

	slog.Info("Starting server", "url", fmt.Sprintf("http://%s", config.GetServer()))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sc := echo.StartConfig{
		Address:         config.GetServer(),
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: 10 * time.Second,
		BeforeServeFunc: func(s *http.Server) error {
			s.ReadHeaderTimeout = 10 * time.Second
			s.ReadTimeout = 30 * time.Second
			s.IdleTimeout = 120 * time.Second
			return nil
		},
	}
	if err := sc.Start(ctx, e); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
