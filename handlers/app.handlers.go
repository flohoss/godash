package handlers

import (
	"github.com/labstack/echo/v4"
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

func NewAppHandler(env *env.Config, s SystemService, w WeatherService, b BookmarkService) *AppHandler {

	return &AppHandler{
		env:             env,
		SystemService:   s,
		WeatherService:  w,
		BookmarkService: b,
	}
}

type AppHandler struct {
	env             *env.Config
	SystemService   SystemService
	WeatherService  WeatherService
	BookmarkService BookmarkService
}

func (bh *AppHandler) appHandler(c echo.Context) error {
	bookmarks := bh.BookmarkService.GetAllBookmarks()
	staticSystem := bh.SystemService.GetStaticInformation()
	liveSystem := bh.SystemService.GetLiveInformation()
	weather := bh.WeatherService.GetCurrentWeather()

	titlePage := bh.env.Title

	return renderView(c, home.HomeIndex(titlePage, bh.env.Version, home.Home(titlePage, bookmarks, staticSystem, liveSystem, weather)))
}
