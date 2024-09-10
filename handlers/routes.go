package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/r3labs/sse/v2"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	router.HandleFunc("GET /sse", sse.ServeHTTP)

	fsAssets := http.FileServer(http.Dir("assets"))
	router.Handle("GET /assets/", http.StripPrefix("/assets/", fsAssets))

	icons := http.FileServer(http.Dir("storage/icons"))
	router.Handle("GET /icons/", http.StripPrefix("/icons/", icons))

	state := func() string {
		return uuid.New().String()
	}
	router.Handle("GET /login", rp.AuthURLHandler(state, authHandler.provider, authHandler.urlOptions...))
	router.Handle("GET /auch/callback", rp.CodeExchangeHandler(authHandler.marshalToken, authHandler.provider))

	router.HandleFunc("GET /", authMiddleware(http.HandlerFunc(appHandler.appHandler), authHandler))
}

func authMiddleware(next http.Handler, authHandler *AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		clains, err := rp.VerifyTokens(ctx, authHandler.provider)
		next.ServeHTTP(w, r)
	}
}
