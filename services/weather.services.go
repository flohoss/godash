package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/pkg/meteo"
	"github.com/r3labs/sse/v2"
)

type WeatherService struct {
	weather      []Day
	hourly       []Hour
	sse          *sse.Server
	mu           sync.RWMutex
	lastResponse *meteo.WeatherResponse
}

type Day struct {
	Name           string `json:"name"`
	TemperatureMax string `json:"temperature_max"`
	TemperatureMin string `json:"temperature_min"`
	Icon           string `json:"icon"`
	More           More   `json:"more"`
}

type Hour struct {
	Time        string `json:"time"`
	Temperature string `json:"temperature"`
	Icon        string `json:"icon"`
	WindSpeed   string `json:"wind_speed"`
	PrecipProb  string `json:"precip_prob"`
}

type More struct {
	CurrentTemperature  string `json:"current_temperature"`
	ApparentTemperature string `json:"apparent_temperature"`
	Humidity            string `json:"humidity"`
	WindSpeed           string `json:"wind_speed"`
	Sunrise             string `json:"sunrise"`
	Sunset              string `json:"sunset"`
}

func NewWeatherService(sse *sse.Server) *WeatherService {
	w := &WeatherService{sse: sse}
	sse.CreateStream("weather")

	if err := w.fetchAndPublish(); err != nil {
		slog.Error("Failed initial weather fetch", "error", err)
		w.weather = []Day{{
			Name:           "Loading...",
			TemperatureMax: "--",
			TemperatureMin: "--",
			Icon:           "icon-[bi--cloud-fill]",
			More:           More{},
		}}
	}

	RegisterSnapshot("weather", w.publishSnapshot)

	go w.collect()
	return w
}

func (w *WeatherService) publishSnapshot() {
	w.mu.RLock()
	weather := w.weather
	hourly := w.hourly
	w.mu.RUnlock()
	if len(weather) == 0 {
		return
	}
	w.publishCurrent(weather[0])
	w.publishForecast(weather)
	w.publishHourly(hourly)
}

func (w *WeatherService) publishCurrent(day Day) {
	data, err := json.Marshal(day)
	if err != nil {
		return
	}
	w.sse.Publish("weather", &sse.Event{Event: []byte("current"), Data: append([]byte(nil), data...)})
}

func (w *WeatherService) publishForecast(days []Day) {
	data, err := json.Marshal(days)
	if err != nil {
		return
	}
	w.sse.Publish("weather", &sse.Event{Event: []byte("forecast"), Data: append([]byte(nil), data...)})
}

func (w *WeatherService) publishHourly(hours []Hour) {
	data, err := json.Marshal(hours)
	if err != nil {
		return
	}
	w.sse.Publish("weather", &sse.Event{Event: []byte("hourly"), Data: append([]byte(nil), data...)})
}

func (w *WeatherService) GetCurrentWeather() []Day {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.weather
}

func (w *WeatherService) GetCurrentHourly() []Hour {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.hourly
}

func (w *WeatherService) collect() {
	fetchTicker := time.NewTicker(5 * time.Minute)
	defer fetchTicker.Stop()
	hourlyTicker := time.NewTicker(1 * time.Minute)
	defer hourlyTicker.Stop()

	for {
		select {
		case <-fetchTicker.C:
			if err := w.fetchAndPublish(); err != nil {
				slog.Error("Failed to update weather", "error", err)
			}
		case <-hourlyTicker.C:
			w.recomputeHourly()
		}
	}
}

func (w *WeatherService) recomputeHourly() {
	w.mu.RLock()
	res := w.lastResponse
	w.mu.RUnlock()
	if res == nil {
		return
	}
	hours := buildHourly(res)
	w.mu.Lock()
	w.hourly = hours
	w.mu.Unlock()
	w.publishHourly(hours)
}

func (w *WeatherService) fetchAndPublish() error {
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
		w.recomputeHourly()
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
				WindSpeed:           fmt.Sprintf("%.0f %s", res.Current.WindSpeed10m, res.CurrentUnits.WindSpeed10m),
				Sunrise:             sunrise.Format("15:04"),
				Sunset:              sunset.Format("15:04"),
			}
		}
		newWeather = append(newWeather, day)
	}

	hours := buildHourly(&res)

	w.mu.Lock()
	w.weather = newWeather
	w.hourly = hours
	w.lastResponse = &res
	w.mu.Unlock()

	w.publishCurrent(newWeather[0])
	w.publishForecast(newWeather)
	w.publishHourly(hours)

	return nil
}

func buildHourly(res *meteo.WeatherResponse) []Hour {
	if len(res.Hourly.Time) == 0 {
		return []Hour{}
	}
	loc, err := time.LoadLocation(config.GetTimeZone())
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	start := now.Truncate(time.Hour).Add(time.Hour)
	hours := make([]Hour, 0, 8)
	for t := start; len(hours) < 8; t = t.Add(time.Hour) {
		idx := nearestHourIndex(res, t)
		if idx < 0 {
			continue
		}
		isDay := res.Hourly.IsDay[idx] == 1
		hours = append(hours, Hour{
			Time:        t.Format("15:04"),
			Temperature: fmt.Sprintf("%.0f %s", res.Hourly.Temperature2m[idx], res.HourlyUnits.Temperature2m),
			Icon:        getIcon(res.Hourly.WeatherCode[idx], isDay),
			WindSpeed:   fmt.Sprintf("%.0f %s", res.Hourly.WindSpeed10m[idx], res.HourlyUnits.WindSpeed10m),
			PrecipProb:  fmt.Sprintf("%d%%", res.Hourly.PrecipitationProbability[idx]),
		})
	}
	return hours
}

func nearestHourIndex(res *meteo.WeatherResponse, target time.Time) int {
	bestIdx := -1
	bestDelta := time.Duration(1<<63 - 1)
	for i, ts := range res.Hourly.Time {
		t, err := time.ParseInLocation("2006-01-02T15:04", ts, target.Location())
		if err != nil {
			continue
		}
		delta := t.Sub(target)
		if delta < 0 {
			delta = -delta
		}
		if delta < bestDelta {
			bestDelta = delta
			bestIdx = i
		}
	}
	return bestIdx
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
		prev.Current.ApparentTemperature != newRes.Current.ApparentTemperature ||
		prev.Current.WindSpeed10m != newRes.Current.WindSpeed10m {
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
