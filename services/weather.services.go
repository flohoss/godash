package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/r3labs/sse/v2"
	"gitlab.unjx.de/flohoss/godash/config"
)

func NewWeatherService(sse *sse.Server) *WeatherService {
	var w = WeatherService{sse: sse}
	go w.updateWeather(time.Second * 90)
	return &w
}

func (w *WeatherService) GetCurrentWeather() *OpenWeather {
	return &w.CurrentWeather
}

func (w *WeatherService) copyWeatherValues(weatherResp *OpenWeatherApiResponse) {
	myTime := time.Unix(weatherResp.Sys.Sunrise, 0)
	w.CurrentWeather.Sunrise = myTime.Format("15:04")
	myTime = time.Unix(weatherResp.Sys.Sunset, 0)
	w.CurrentWeather.Sunset = myTime.Format("15:04")
	w.CurrentWeather.Icon = weatherResp.Weather[0].Icon
	w.CurrentWeather.Temp = int(math.Round(weatherResp.Main.Temp))
	w.CurrentWeather.Description = weatherResp.Weather[0].Description
	w.CurrentWeather.Humidity = weatherResp.Main.Humidity
}

func (w *WeatherService) updateWeather(interval time.Duration) {
	w.sse.CreateStream("weather")

	for {
		settings := config.GetWeatherSettings()
		if settings.Key == "" {
			return
		}

		if settings.Units == "imperial" {
			w.CurrentWeather.Units = "°F"
		} else {
			w.CurrentWeather.Units = "°C"
		}

		location := config.GetLocation()

		resp, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=%s&lang=%s",
			location.Latitude,
			location.Longitude,
			settings.Key,
			settings.Units,
			settings.Language))
		if err != nil || resp.StatusCode != 200 {
			slog.Error("weather cannot be updated, please check WEATHER_KEY")
		} else {
			body, _ := io.ReadAll(resp.Body)
			var weatherResponse OpenWeatherApiResponse
			err = json.Unmarshal(body, &weatherResponse)
			if err != nil {
				slog.Error("weather cannot be processed")
			} else {
				w.copyWeatherValues(&weatherResponse)
			}
			resp.Body.Close()
			json, _ := json.Marshal(w.CurrentWeather)
			w.sse.Publish("weather", &sse.Event{Data: json})
		}
		time.Sleep(interval)
	}
}
