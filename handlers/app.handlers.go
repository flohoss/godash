package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/logto-io/go/client"
	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/services"
	"gitlab.unjx.de/flohoss/godash/views/home"
)

type BookmarkService interface {
	GetAllBookmarks() *services.Bookmarks
}

type SystemService interface {
	GetLiveInformation() *services.LiveInformation
	GetStaticInformation() *services.StaticInformation
}

type WeatherService interface {
	GetCurrentWeather() *services.OpenWeather
}

func NewAppHandler(env *env.Config, authHandler *AuthHandler, s SystemService, w WeatherService, b BookmarkService) *AppHandler {
	return &AppHandler{
		env:             env,
		authHandler:     authHandler,
		systemService:   s,
		weatherService:  w,
		bookmarkService: b,
	}
}

type AppHandler struct {
	env             *env.Config
	authHandler     *AuthHandler
	systemService   SystemService
	weatherService  WeatherService
	bookmarkService BookmarkService
}

func (bh *AppHandler) appHandler(c echo.Context) error {
	bookmarks := bh.bookmarkService.GetAllBookmarks()
	staticSystem := bh.systemService.GetStaticInformation()
	liveSystem := bh.systemService.GetLiveInformation()
	weather := bh.weatherService.GetCurrentWeather()

	logtoClient := client.NewLogtoClient(
		bh.authHandler.logtoConfig,
		NewSessionStorage(c),
	)
	user, _ := logtoClient.FetchUserInfo()

	titlePage := bh.env.Title

	return renderView(c, home.HomeIndex(titlePage, bh.env.Version, home.Home(titlePage, user, bookmarks, staticSystem, liveSystem, weather)))
}
