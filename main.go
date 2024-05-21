package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

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

	lis := fmt.Sprintf(":%d", env.Port)
	slog.Info("server listening, press ctrl+c to stop", "addr", "http://localhost"+lis)
	err = http.ListenAndServe(lis, router)
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server terminated", "error", err)
		os.Exit(1)
	}
}
