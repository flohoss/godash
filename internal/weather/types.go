package weather

import (
	"github.com/r3labs/sse/v2"
)

type Weather struct {
	CurrentWeather OpenWeather
	sse            *sse.Server
	config         config
}

type config struct {
	Latitude  float32 `env:"LOCATION_LATITUDE" envDefault:"48.780331609463815"`
	Longitude float32 `env:"LOCATION_LONGITUDE" envDefault:"9.177968320179422"`
	Key       string  `env:"WEATHER_KEY" envDefault:""`
	Units     string  `env:"WEATHER_UNITS" envDefault:"metric"`
	Lang      string  `env:"WEATHER_LANG" envDefault:"en"`
	Digits    bool    `env:"WEATHER_DIGITS" envDefault:"true"`
}

type OpenWeather struct {
	Icon        string  `json:"icon"`
	Temp        float64 `json:"temp"`
	Description string  `json:"description"`
	Humidity    uint8   `json:"humidity"`
	Sunrise     string  `json:"sunrise"`
	Sunset      string  `json:"sunset"`
	Units       string  `json:"units"`
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
