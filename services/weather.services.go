package services

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/pkg/meteo"
	"github.com/r3labs/sse/v2"
)

type WeatherService struct {
	weather        []Day
	sse            *sse.Server
	mu             sync.RWMutex
	renderCurrent  func(Day) templ.Component
	renderForecast func([]Day) templ.Component
	lastResponse   *meteo.WeatherResponse
}

type Day struct {
	Name           string `json:"name"`
	TemperatureMax string `json:"temperature_max"`
	TemperatureMin string `json:"temperature_min"`
	Icon           string `json:"icon"`
	More           More   `json:"more"`
}

type More struct {
	CurrentTemperature  string `json:"current_temperature"`
	ApparentTemperature string `json:"apparent_temperature"`
	Humidity            string `json:"humidity"`
	Sunrise             string `json:"sunrise"`
	Sunset              string `json:"sunset"`
}

func NewWeatherService(sse *sse.Server, renderCurrent func(Day) templ.Component, renderForecast func([]Day) templ.Component) *WeatherService {
	w := &WeatherService{
		sse:            sse,
		renderCurrent:  renderCurrent,
		renderForecast: renderForecast,
	}
	sse.CreateStream("weather")

	var currentBuf, forecastBuf bytes.Buffer
	if err := w.fetchAndPublish(&currentBuf, &forecastBuf); err != nil {
		slog.Error("Failed initial weather fetch", "error", err)
		w.weather = []Day{{
			Name:           "Loading...",
			TemperatureMax: "--",
			TemperatureMin: "--",
			Icon:           "icon-[bi--cloud-fill]",
			More:           More{},
		}}
	}

	go w.collect()
	return w
}

func (w *WeatherService) GetCurrentWeather() []Day {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.weather
}

func (w *WeatherService) collect() {
	ticker := time.NewTicker(90 * time.Second)
	defer ticker.Stop()

	var currentBuf, forecastBuf bytes.Buffer

	for range ticker.C {
		if err := w.fetchAndPublish(&currentBuf, &forecastBuf); err != nil {
			slog.Error("Failed to update weather", "error", err)
		}
	}
}

func (w *WeatherService) fetchAndPublish(currentBuf, forecastBuf *bytes.Buffer) error {
	settings := config.GetWeatherSettings()
	res, err := meteo.GetWeather(meteo.Options{
		Latitude:  settings.Latitude,
		Longitude: settings.Longitude,
		TimeZone:  config.GetTimeZone(),
		Units:     settings.Units,
	})
	if err != nil {
		return err
	}

	w.mu.RLock()
	hasChanged := w.lastResponse == nil || w.hasResponseChanged(&res)
	w.mu.RUnlock()

	if !hasChanged {
		return nil
	}

	newWeather := []Day{}
	for i, t := range res.Daily.Time {
		t, _ := time.Parse("2006-01-02", t)
		day := Day{
			Name:           t.Format("Mon 02 Jan"),
			TemperatureMax: fmt.Sprintf("%.1f %s", res.Daily.TemperatureMax[i], res.DailyUnits.TemperatureMax),
			TemperatureMin: fmt.Sprintf("%.1f %s", res.Daily.TemperatureMin[i], res.DailyUnits.TemperatureMin),
			Icon:           getIcon(res.Daily.WeatherCode[i], res.Current.IsDay == 1),
		}
		if i == 0 {
			sunrise, _ := time.Parse("2006-01-02T15:04", res.Daily.Sunrise[0])
			sunset, _ := time.Parse("2006-01-02T15:04", res.Daily.Sunset[0])
			day.Icon = getIcon(res.Current.WeatherCode, res.Current.IsDay == 1)
			day.More = More{
				CurrentTemperature:  fmt.Sprintf("%.1f %s", res.Current.Temperature2m, res.CurrentUnits.Temperature2m),
				ApparentTemperature: fmt.Sprintf("%.1f %s", res.Current.ApparentTemperature, res.CurrentUnits.ApparentTemperature),
				Humidity:            fmt.Sprintf("%d %s", res.Current.RelativeHumidity, res.CurrentUnits.RelativeHumidity),
				Sunrise:             sunrise.Format("15:04"),
				Sunset:              sunset.Format("15:04"),
			}
		}
		newWeather = append(newWeather, day)
	}

	w.mu.Lock()
	w.weather = newWeather
	w.lastResponse = &res
	w.mu.Unlock()

	currentBuf.Reset()
	if err := w.renderCurrent(newWeather[0]).Render(context.Background(), currentBuf); err != nil {
		return err
	}
	w.sse.Publish("weather", &sse.Event{Event: []byte("current"), Data: currentBuf.Bytes()})

	forecastBuf.Reset()
	if err := w.renderForecast(newWeather).Render(context.Background(), forecastBuf); err != nil {
		return err
	}
	w.sse.Publish("weather", &sse.Event{Event: []byte("forecast"), Data: forecastBuf.Bytes()})

	return nil
}

func (w *WeatherService) hasResponseChanged(newRes *meteo.WeatherResponse) bool {
	if w.lastResponse == nil {
		return true
	}

	prev := w.lastResponse

	if prev.Current.Temperature2m != newRes.Current.Temperature2m ||
		prev.Current.WeatherCode != newRes.Current.WeatherCode ||
		prev.Current.IsDay != newRes.Current.IsDay ||
		prev.Current.RelativeHumidity != newRes.Current.RelativeHumidity ||
		prev.Current.ApparentTemperature != newRes.Current.ApparentTemperature {
		return true
	}

	if len(prev.Daily.TemperatureMax) > 0 && len(newRes.Daily.TemperatureMax) > 0 {
		if prev.Daily.TemperatureMax[0] != newRes.Daily.TemperatureMax[0] ||
			prev.Daily.TemperatureMin[0] != newRes.Daily.TemperatureMin[0] {
			return true
		}
	}

	if len(prev.Daily.TemperatureMax) > 1 && len(newRes.Daily.TemperatureMax) > 1 {
		if prev.Daily.TemperatureMax[1] != newRes.Daily.TemperatureMax[1] ||
			prev.Daily.TemperatureMin[1] != newRes.Daily.TemperatureMin[1] ||
			prev.Daily.WeatherCode[1] != newRes.Daily.WeatherCode[1] {
			return true
		}
	}

	return false
}

func getIcon(code int, isDay bool) string {
	switch code {
	case 0:
		if isDay {
			return "icon-[bi--sun-fill]"
		}
		return "icon-[bi--moon-fill]"

	case 1, 2:
		if isDay {
			return "icon-[bi--cloud-sun-fill]"
		}
		return "icon-[bi--cloud-moon-fill]"

	case 3:
		return "icon-[bi--cloud-fill]"

	case 45, 48:
		return "icon-[bi--cloud-fog2-fill]"

	case 51, 53, 55, 56, 57, 61, 66, 67, 80:
		return "icon-[bi--cloud-drizzle-fill]"

	case 63, 65, 81:
		return "icon-[bi--cloud-rain-heavy-fill]"

	case 71, 73, 75, 77, 85, 86:
		return "icon-[bi--cloud-snow-fill]"

	case 82, 95, 96, 99:
		return "icon-[bi--cloud-lightning-rain-fill]"

	default:
		return "icon-[bi--cloud-fill]"
	}
}
