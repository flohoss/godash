package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/r3labs/sse/v2"

	"gitlab.unjx.de/flohoss/godash/handlers"
	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/internal/logger"
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

	parsedUrl, _ := url.Parse(env.PublicUrl)
	secret := []byte(env.SessionKey)
	if len(secret) == 0 {
		secret = securecookie.GenerateRandomKey(32)
	}
	store := sessions.NewCookieStore(secret)
	store.Options = &sessions.Options{
		Domain:      parsedUrl.Hostname(),
		MaxAge:      86400 * 30,
		Secure:      parsedUrl.Scheme == "https",
		HttpOnly:    true,
		Partitioned: true,
		SameSite:    http.SameSiteLaxMode,
	}

	authHandler := handlers.NewAuthHandler(env, store)
	appHandler := handlers.NewAppHandler(env, store, s, w, b)
	handlers.SetupRoutes(router, sse, appHandler, authHandler)

	slog.Info("server listening, press ctrl+c to stop", "addr", env.PublicUrl)
	err = http.ListenAndServe(fmt.Sprintf(":%d", env.Port), router)
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server terminated", "error", err)
		os.Exit(1)
	}
}
