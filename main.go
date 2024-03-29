package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r3labs/sse/v2"

	"gitlab.unjx.de/flohoss/godash/handlers"
	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/services"

	"github.com/glebarez/sqlite"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
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

	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "sse") || strings.Contains(c.Path(), "sign")
		},
	}))
	if env.SSOEndpoint != "" {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			slog.Error("cannot connect to database", "err", err)
			os.Exit(1)
		}
		store := gormstore.New(db, []byte(env.SessionKey))
		quit := make(chan struct{})
		go store.PeriodicCleanup(1*time.Hour, quit)
		e.Use(session.Middleware(store))
	}

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
