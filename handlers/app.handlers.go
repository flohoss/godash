package handlers

import (
	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/godash/config"
	"gitlab.unjx.de/flohoss/godash/services"
	"gitlab.unjx.de/flohoss/godash/views"
)

type SystemService interface {
	GetLiveInformation() *services.LiveInformation
	GetStaticInformation() *services.StaticInformation
}

type WeatherService interface {
	GetCurrentWeather() *services.OpenWeather
}

func NewAppHandler(s SystemService, w WeatherService) *AppHandler {
	return &AppHandler{
		systemService:  s,
		weatherService: w,
	}
}

type AppHandler struct {
	systemService  SystemService
	weatherService WeatherService
}

func (bh *AppHandler) handleIndex(ctx echo.Context) error {
	staticSystem := bh.systemService.GetStaticInformation()
	liveSystem := bh.systemService.GetLiveInformation()
	weather := bh.weatherService.GetCurrentWeather()

	return render(ctx, views.HomeIndex(config.GetTitle(), views.Home(config.GetApplications(), config.GetLinks(), staticSystem, liveSystem, weather)))
}
