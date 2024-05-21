package handlers

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler) {
	router.HandleFunc("GET /sse", sse.ServeHTTP)

	fsAssets := http.FileServer(http.Dir("assets"))
	router.Handle("GET /assets/", http.StripPrefix("/assets/", fsAssets))

	fsIcons := http.FileServer(http.Dir("storage/icons"))
	router.Handle("GET /storage/icons/", http.StripPrefix("/storage/icons/", fsIcons))

	router.HandleFunc("GET /", appHandler.appHandler)
}
