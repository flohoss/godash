package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/pkg/meteo"
	"github.com/r3labs/sse/v2"
)

type WeatherService struct {
	weather []Day
	sse     *sse.Server
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

func NewWeatherService(sse *sse.Server) *WeatherService {
	var w = WeatherService{sse: sse}
	w.sse.CreateStream("weather")
	interval := time.Second * 90
	w.updateWeather(interval)
	go func() {
		for {
			if err := w.updateWeather(interval); err != nil {
				slog.Error("Failed to update weather", "error", err)
			}
			time.Sleep(interval)
		}
	}()
	return &w
}

func (w *WeatherService) GetCurrentWeather() []Day {
	return w.weather
}

func (w *WeatherService) updateWeather(interval time.Duration) error {
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

	json, _ := json.Marshal(newWeather)
	w.sse.Publish("weather", &sse.Event{Data: json})
	w.weather = newWeather
	return nil
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
