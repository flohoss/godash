package handlers

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	router.HandleFunc("GET /sse", sse.ServeHTTP)

	fsAssets := http.FileServer(http.Dir("assets"))
	router.Handle("GET /assets/", http.StripPrefix("/assets/", fsAssets))

	icons := http.FileServer(http.Dir("storage/icons"))
	router.Handle("GET /icons/", http.StripPrefix("/icons/", icons))

	router.HandleFunc("GET /login", authHandler.handleAuth)
	router.HandleFunc("GET /auch/callback", authHandler.handleCallback)

	router.Handle("GET /", authHandler.AuthMiddleware(http.HandlerFunc(appHandler.appHandler)))
}
