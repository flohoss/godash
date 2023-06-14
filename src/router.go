package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (g *goDash) setupRouter() {
	g.router.GET("/", g.index)
	g.router.GET("/robots.txt", robots)

	g.router.GET("/sse", echo.WrapHandler(http.HandlerFunc(g.sse.ServeHTTP)))

	static := g.router.Group("/static", longCacheLifetime)
	static.Static("/", "static")

	storage := g.router.Group("/storage", longCacheLifetime)
	storage.Static("/icons", "storage/icons")

	g.router.RouteNotFound("/*", redirectHome)
}
