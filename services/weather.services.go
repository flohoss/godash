package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/r3labs/sse/v2"
	"gitlab.unjx.de/flohoss/godash/config"
	"gitlab.unjx.de/flohoss/godash/pkg/meteo"
)

type WeatherService struct {
	weather []Day
	sse     *sse.Server
}

type Day struct {
	Name           string `json:"day"`
	PrimaryValue   string `json:"primary_value"`
	SecondaryValue string `json:"secondary_value"`
	Icon           string `json:"icon"`
	Sunrise        string `json:"sunrise"`
	Sunset         string `json:"sunset"`
}

func NewWeatherService(sse *sse.Server) *WeatherService {
	var w = WeatherService{sse: sse}
	go w.updateWeather(time.Second * 90)
	return &w
}

func (w *WeatherService) GetCurrentWeather() []Day {
	return w.weather
}

func (w *WeatherService) updateWeather(interval time.Duration) {
	w.sse.CreateStream("weather")

	for {
		settings := config.GetWeatherSettings()
		res, err := meteo.GetWeather(meteo.Options{
			Latitude:  settings.Latitude,
			Longitude: settings.Longitude,
			TimeZone:  config.GetTimeZone(),
			Units:     settings.Units,
		})
		if err != nil {
			slog.Error("Failed to get weather", "error", err)
			time.Sleep(interval)
			continue
		}
		newWeather := []Day{}
		newWeather = append(newWeather, Day{
			Name:           time.Now().Format("Mon 02 Jan"),
			PrimaryValue:   fmt.Sprintf("%.1f %s", res.Current.Temperature2m, res.CurrentUnits.Temperature2m),
			SecondaryValue: fmt.Sprintf("%d %s", res.Current.RelativeHumidity, res.CurrentUnits.RelativeHumidity),
			Icon:           getIcon(res.Current.WeatherCode, res.Current.IsDay == 1),
			Sunrise:        res.Daily.Sunrise[0],
			Sunset:         res.Daily.Sunset[0],
		})
		sunrise, _ := time.Parse("2006-01-02T15:04", res.Daily.Sunrise[0])
		sunset, _ := time.Parse("2006-01-02T15:04", res.Daily.Sunset[0])
		newWeather = append(newWeather, Day{
			Name:           time.Now().Format("Mon 02 Jan"),
			PrimaryValue:   fmt.Sprintf("%.1f %s", res.Current.Temperature2m, res.CurrentUnits.Temperature2m),
			SecondaryValue: fmt.Sprintf("%d %s", res.Current.RelativeHumidity, res.CurrentUnits.RelativeHumidity),
			Icon:           getIcon(res.Current.WeatherCode, res.Current.IsDay == 1),
			Sunrise:        sunrise.Format("15:04"),
			Sunset:         sunset.Format("15:04"),
		})
		for i := 1; i < 6; i++ {
			date, _ := time.Parse("2006-01-02", res.Daily.Time[i])
			sunrise, _ := time.Parse("2006-01-02T15:04", res.Daily.Sunrise[i])
			sunset, _ := time.Parse("2006-01-02T15:04", res.Daily.Sunset[i])
			newWeather = append(newWeather, Day{
				Name:           date.Format("Mon 02 Jan"),
				PrimaryValue:   fmt.Sprintf("%.1f %s", res.Daily.TemperatureMax[i], res.DailyUnits.TemperatureMax),
				SecondaryValue: fmt.Sprintf("%.1f %s", res.Daily.TemperatureMin[i], res.DailyUnits.TemperatureMin),
				Icon:           getIcon(res.Daily.WeatherCode[i], true),
				Sunrise:        sunrise.Format("15:04"),
				Sunset:         sunset.Format("15:04"),
			})
		}

		w.weather = newWeather
		json, _ := json.Marshal(newWeather)
		w.sse.Publish("weather", &sse.Event{Data: json})
		time.Sleep(interval)
	}
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
