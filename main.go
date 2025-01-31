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
	"gitlab.unjx.de/flohoss/godash/pkg/logger"
	"gitlab.unjx.de/flohoss/godash/services"
)

func main() {
	slog.SetDefault(logger.NewLogger())

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

	slog.Info("server listening, press ctrl+c to stop", "addr", env.PublicUrl)
	err = http.ListenAndServe(fmt.Sprintf(":%d", env.Port), router)
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server terminated", "error", err)
		os.Exit(1)
	}
}
