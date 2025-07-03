package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	router := http.NewServeMux()
	sse := sse.New()
	sse.AutoReplay = false

	s := services.NewSystemService(sse)
	w := services.NewWeatherService(sse, env)
	b := services.NewBookmarkService()

	appHandler := handlers.NewAppHandler(env, s, w, b)
	handlers.SetupRoutes(router, sse, appHandler)

	slog.Info("server starting", "addr", env.PublicUrl)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", env.Port), router); err != nil && err != http.ErrServerClosed {
			slog.Error("shutting down the server")
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Received shutdown signal. Exiting immediately.")
}
