package handlers

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	if authHandler.sessionManager != nil {
		router.HandleFunc("GET /sign-in", authHandler.signInHandler)
		router.HandleFunc("GET /sign-in-callback", authHandler.signInCallbackHandler)
		router.HandleFunc("GET /sign-out", authHandler.signOutHandler)
	}
	router.Handle("GET /sse", authHandler.authRequired(http.HandlerFunc(sse.ServeHTTP)))

	fsAssets := http.FileServer(http.Dir("assets"))
	router.Handle("GET /assets/", authHandler.authRequired(http.StripPrefix("/assets/", fsAssets)))

	fsIcons := http.FileServer(http.Dir("storage/icons"))
	router.Handle("GET /storage/icons/", authHandler.authRequired(http.StripPrefix("/storage/icons/", fsIcons)))

	router.Handle("GET /", authHandler.authRequired(http.HandlerFunc(appHandler.appHandler)))
}
