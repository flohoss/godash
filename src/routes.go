package main

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{}
)

func (g *goDash) index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Title":   g.config.Title,
		"Weather": g.info.weather.CurrentWeather,
		"Parsed":  g.info.bookmarks.Parsed,
		"System":  g.info.system,
	})
}

func robots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /")
}

func redirectHome(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/")
}
