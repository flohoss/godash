package handlers

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	if authHandler.oidc != nil {
		router.Handle("/auth/", authHandler.oidc)
	}
	router.Handle("/sse", authHandler.mw.RequireAuthentication()(http.HandlerFunc(sse.ServeHTTP)))

	fsAssets := http.FileServer(http.Dir("assets"))
	router.Handle("/assets/", http.StripPrefix("/assets/", fsAssets))

	fsIcons := http.FileServer(http.Dir("storage/icons"))
	router.Handle("/storage/icons/", http.StripPrefix("/storage/icons/", fsIcons))

	router.Handle("/", authHandler.mw.RequireAuthentication()(http.HandlerFunc(appHandler.appHandler)))
}
