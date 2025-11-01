package handlers

import (
	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/services"
	"github.com/flohoss/godash/views"
	"github.com/labstack/echo/v4"
)

type SystemService interface {
	GetBuffer() *services.Buffer
	GetStatic() *services.Static
}

type WeatherService interface {
	GetCurrentWeather() []services.Day
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
	buffer := bh.systemService.GetBuffer()
	static := bh.systemService.GetStatic()
	weather := bh.weatherService.GetCurrentWeather()

	return render(ctx, views.Home(config.GetTitle(), config.GetApplications(), config.GetLinks(), buffer, static, weather))
}
