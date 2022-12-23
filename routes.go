package main

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"godash/hub"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{}
)

func (g *goDash) homePage(c echo.Context) error {
	return c.Render(http.StatusOK, "index", map[string]interface{}{
		"Title":   g.config.Title,
		"Weather": g.info.weather.CurrentWeather,
		"Parsed":  g.info.bookmarks.Parsed,
		"System":  g.info.system,
	})
}

func (g *goDash) ws(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil
	}
	messageChan := make(hub.NotifierChan)
	g.hub.NewClients <- messageChan
	defer func() {
		g.hub.ClosingClients <- messageChan
		ws.Close()
	}()

	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	for {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				_ = ws.WriteMessage(websocket.CloseMessage, []byte{})
			}
			err := ws.WriteJSON(msg)
			if err != nil {
				return nil
			}
		}
	}
}

func robotsHandler(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /")
}

func redirectHome(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
