package handlers

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	router.Handle("GET /sse", authHandler.AuthMiddleware(http.HandlerFunc(sse.ServeHTTP)))

	fsAssets := http.FileServer(http.Dir("assets"))
	router.Handle("GET /assets/", authHandler.AuthMiddleware(http.StripPrefix("/assets/", fsAssets)))

	icons := http.FileServer(http.Dir("storage/icons"))
	router.Handle("GET /icons/", authHandler.AuthMiddleware(http.StripPrefix("/icons/", icons)))

	router.HandleFunc("GET /logout", authHandler.handleLogout)
	router.HandleFunc("GET /callback", authHandler.handleCallback)

	router.Handle("GET /", authHandler.AuthMiddleware(http.HandlerFunc(appHandler.appHandler)))
}
