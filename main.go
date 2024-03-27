package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r3labs/sse/v2"
	"gitlab.unjx.de/flohoss/godash/handlers"
	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/services"
)

func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = true

	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "sse") || strings.Contains(c.Path(), "auth")
		},
	}))

	sse := sse.New()
	sse.AutoReplay = false

	s := services.NewSystemService(sse)
	w := services.NewWeatherService(sse, env)
	b := services.NewBookmarkService()

	authHandler := handlers.NewAuthHandler(env)
	appHandler := handlers.NewAppHandler(env, authHandler, s, w, b)
	handlers.SetupRoutes(e, sse, appHandler, authHandler)

	slog.Info("starting server", "url", env.PublicUrl)
	if err := e.Start(fmt.Sprintf("0.0.0.0:%d", env.Port)); err != http.ErrServerClosed {
		slog.Error("cannot start server", "err", err)
		os.Exit(1)
	}
}
