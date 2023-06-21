package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/caarlos0/env/v8"
	"github.com/r3labs/sse/v2"
	"go.uber.org/zap"
)

func NewWeatherService(sse *sse.Server) *Weather {
	var w = Weather{sse: sse}
	if err := env.Parse(&w.config); err != nil {
		panic(err)
	}
	if w.config.Key != "" {
		w.setWeatherUnits()
		go w.updateWeather(time.Second * 90)
	}
	return &w
}

func (w *Weather) setWeatherUnits() {
	if w.config.Units == "imperial" {
		w.CurrentWeather.Units = "°F"
	} else {
		w.CurrentWeather.Units = "°C"
	}
}

func (w *Weather) copyWeatherValues(weatherResp *OpenWeatherApiResponse) {
	myTime := time.Unix(weatherResp.Sys.Sunrise, 0)
	w.CurrentWeather.Sunrise = myTime.Format("15:04")
	myTime = time.Unix(weatherResp.Sys.Sunset, 0)
	w.CurrentWeather.Sunset = myTime.Format("15:04")
	w.CurrentWeather.Icon = weatherResp.Weather[0].Icon
	if w.config.Digits {
		w.CurrentWeather.Temp = weatherResp.Main.Temp
	} else {
		w.CurrentWeather.Temp = math.Round(weatherResp.Main.Temp)
	}
	w.CurrentWeather.Description = weatherResp.Weather[0].Description
	w.CurrentWeather.Humidity = weatherResp.Main.Humidity
}

func (w *Weather) updateWeather(interval time.Duration) {
	var weatherResponse OpenWeatherApiResponse
	for {
		resp, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=%s&lang=%s",
			w.config.Latitude,
			w.config.Longitude,
			w.config.Key,
			w.config.Units,
			w.config.Lang))
		if err != nil || resp.StatusCode != 200 {
			zap.L().Error("weather cannot be updated, please check WEATHER_KEY")
		} else {
			body, _ := io.ReadAll(resp.Body)
			err = json.Unmarshal(body, &weatherResponse)
			if err != nil {
				zap.L().Error("weather cannot be processed")
			} else {
				w.copyWeatherValues(&weatherResponse)
				zap.L().Debug("weather updated", zap.Float64("temp", w.CurrentWeather.Temp))
			}
			resp.Body.Close()
			json, _ := json.Marshal(w.CurrentWeather)
			w.sse.Publish("weather", &sse.Event{Data: json})
		}
		time.Sleep(interval)
	}
}
