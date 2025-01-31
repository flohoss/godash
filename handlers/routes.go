package handlers

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

func SetupRoutes(router *http.ServeMux, sse *sse.Server, appHandler *AppHandler) {
	router.Handle("GET /sse", http.HandlerFunc(sse.ServeHTTP))

	fsAssets := http.FileServer(http.Dir("assets"))
	router.Handle("GET /assets/", http.StripPrefix("/assets/", fsAssets))

	icons := http.FileServer(http.Dir("storage/icons"))
	router.Handle("GET /icons/", http.StripPrefix("/icons/", icons))

	router.Handle("GET /", http.HandlerFunc(appHandler.appHandler))
}
