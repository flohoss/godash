package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"gitlab.unjx.de/flohoss/godash/internal/controller"
	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/internal/logger"
	"gitlab.unjx.de/flohoss/godash/internal/router"
)

func main() {
	env, err := env.Parse()
	if err != nil {
		slog.Error("cannot parse environment variables", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(logger.NewLogger(env.LogLevel))

	r := router.InitRouter()
	c := controller.NewController(env)
	router.SetupRoutes(r, c)

	slog.Info("starting server", "url", fmt.Sprintf("http://localhost:%d", env.Port))
	if err := r.Start(fmt.Sprintf(":%d", env.Port)); err != http.ErrServerClosed {
		slog.Error("cannot start server", "err", err)
		os.Exit(1)
	}
}
