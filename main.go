package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
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

	var sessionManager *scs.SessionManager
	if env.OIDCIssuerUrl != "" {
		sessionManager = scs.New()
		sessionManager.Lifetime = 168 * time.Hour
	}
	router := http.NewServeMux()
	sse := sse.New()
	sse.AutoReplay = false

	s := services.NewSystemService(sse)
	w := services.NewWeatherService(sse, env)
	b := services.NewBookmarkService()

	authHandler := handlers.NewAuthHandler(env, sessionManager)
	appHandler := handlers.NewAppHandler(env, authHandler, s, w, b)
	handlers.SetupRoutes(router, sse, appHandler, authHandler)

	lis := fmt.Sprintf(":%d", env.Port)
	slog.Info("server listening, press ctrl+c to stop", "addr", "http://localhost"+lis)
	if sessionManager != nil {
		err = http.ListenAndServe(lis, sessionManager.LoadAndSave(router))
	} else {
		err = http.ListenAndServe(lis, router)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server terminated", "error", err)
		os.Exit(1)
	}
}
