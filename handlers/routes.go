package handlers

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	if authHandler != nil {
		router.Handle("GET /auth/", authHandler.authenticator)

		router.Handle("GET /sse", authHandler.middleware.RequireAuthentication()(http.HandlerFunc(sse.ServeHTTP)))

		fsAssets := http.FileServer(http.Dir("assets"))
		router.Handle("GET /assets/", authHandler.middleware.RequireAuthentication()(http.StripPrefix("/assets/", fsAssets)))

		fsIcons := http.FileServer(http.Dir("storage/icons"))
		router.Handle("GET /storage/icons/", authHandler.middleware.RequireAuthentication()(http.StripPrefix("/storage/icons/", fsIcons)))

		router.Handle("GET /", authHandler.middleware.RequireAuthentication()(http.HandlerFunc(appHandler.appHandler)))
	} else {
		router.HandleFunc("GET /sse", sse.ServeHTTP)

		fsAssets := http.FileServer(http.Dir("assets"))
		router.Handle("GET /assets/", http.StripPrefix("/assets/", fsAssets))

		fsIcons := http.FileServer(http.Dir("storage/icons"))
		router.Handle("GET /storage/icons/", http.StripPrefix("/storage/icons/", fsIcons))

		router.HandleFunc("GET /", appHandler.appHandler)
	}
}
