package services

import (
	"github.com/r3labs/sse/v2"
)

type WeatherService struct {
	CurrentWeather OpenWeather
	sse            *sse.Server
}

type OpenWeather struct {
	Icon        string `json:"icon"`
	Temp        int    `json:"temp"`
	Description string `json:"description"`
	Humidity    uint8  `json:"humidity"`
	Sunrise     string `json:"sunrise"`
	Sunset      string `json:"sunset"`
	Units       string `json:"units"`
}

type OpenWeatherApiResponse struct {
	Weather []OpenWeatherApiWeather `json:"Weather"`
	Main    OpenWeatherApiMain      `json:"main"`
	Sys     OpenWeatherApiSys       `json:"sys"`
}

type OpenWeatherApiWeather struct {
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type OpenWeatherApiMain struct {
	Temp     float64 `json:"temp"`
	Humidity uint8   `json:"humidity"`
}

type OpenWeatherApiSys struct {
	Sunrise int64 `json:"sunrise"`
	Sunset  int64 `json:"sunset"`
}
