package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flohoss/godash/pkg/meteo"
	"github.com/r3labs/sse/v2"
)

var errTest = errors.New("test error")

func setFetchWeather(t *testing.T, fn func(meteo.Options) (meteo.WeatherResponse, error)) {
	t.Helper()
	old := fetchWeather
	fetchWeather = fn
	t.Cleanup(func() { fetchWeather = old })
}

func testWeatherResponse() meteo.WeatherResponse {
	return meteo.WeatherResponse{
		CurrentUnits: struct {
			Temperature2m       string `json:"temperature_2m"`
			ApparentTemperature string `json:"apparent_temperature"`
			RelativeHumidity    string `json:"relative_humidity_2m"`
			WindSpeed10m        string `json:"wind_speed_10m"`
		}{Temperature2m: "°C", ApparentTemperature: "°C", RelativeHumidity: "%", WindSpeed10m: "km/h"},
		Current: struct {
			Temperature2m       float64 `json:"temperature_2m"`
			ApparentTemperature float64 `json:"apparent_temperature"`
			RelativeHumidity    int     `json:"relative_humidity_2m"`
			WeatherCode         int     `json:"weather_code"`
			IsDay               float64 `json:"is_day"`
			WindSpeed10m        float64 `json:"wind_speed_10m"`
		}{Temperature2m: 21.5, ApparentTemperature: 22.0, RelativeHumidity: 60, WeatherCode: 0, IsDay: 1, WindSpeed10m: 12},
		DailyUnits: struct {
			TemperatureMax string `json:"temperature_2m_max"`
			TemperatureMin string `json:"temperature_2m_min"`
			Sunrise        string `json:"sunrise"`
			Sunset         string `json:"sunset"`
		}{TemperatureMax: "°C", TemperatureMin: "°C"},
		Daily: struct {
			Time           []string  `json:"time"`
			WeatherCode    []int     `json:"weather_code"`
			TemperatureMax []float64 `json:"temperature_2m_max"`
			TemperatureMin []float64 `json:"temperature_2m_min"`
			Sunrise        []string  `json:"sunrise"`
			Sunset         []string  `json:"sunset"`
		}{
			Time:           []string{"2026-08-23", "2026-08-24"},
			WeatherCode:    []int{0, 1},
			TemperatureMax: []float64{25.0, 26.0},
			TemperatureMin: []float64{18.0, 19.0},
			Sunrise:        []string{"2026-08-23T06:00"},
			Sunset:         []string{"2026-08-23T20:00"},
		},
		HourlyUnits: struct {
			Temperature2m            string `json:"temperature_2m"`
			WindSpeed10m             string `json:"wind_speed_10m"`
			PrecipitationProbability string `json:"precipitation_probability"`
		}{Temperature2m: "°C", WindSpeed10m: "km/h", PrecipitationProbability: "%"},
		Hourly: struct {
			Time                     []string  `json:"time"`
			Temperature2m            []float64 `json:"temperature_2m"`
			WeatherCode              []int     `json:"weather_code"`
			IsDay                    []int     `json:"is_day"`
			WindSpeed10m             []float64 `json:"wind_speed_10m"`
			PrecipitationProbability []int     `json:"precipitation_probability"`
		}{
			Time:                     []string{"2026-08-23T12:00"},
			Temperature2m:            []float64{21.0},
			WeatherCode:              []int{0},
			IsDay:                    []int{1},
			WindSpeed10m:             []float64{12},
			PrecipitationProbability: []int{20},
		},
	}
}

// newSSEServer starts an httptest server exposing the SSE handler so tests can
// subscribe and capture published events.
func newSSEServer(t *testing.T) (*sse.Server, *sse.Client) {
	t.Helper()
	s := sse.New()
	s.CreateStream("weather")

	server := httptest.NewServer(http.HandlerFunc(s.ServeHTTP))
	t.Cleanup(server.Close)
	t.Cleanup(s.Close)

	return s, sse.NewClient(server.URL)
}

// collectSSEEvents subscribes to a stream and returns the data of the next n
// events that carry a non-nil payload.
func collectSSEEvents(t *testing.T, client *sse.Client, stream string, n int) []string {
	t.Helper()
	events := make(chan *sse.Event, n)
	go func() {
		_ = client.Subscribe(stream, func(msg *sse.Event) {
			if len(msg.Data) > 0 {
				events <- msg
			}
		})
	}()
	deadline := time.After(3 * time.Second)
	var out []string
	for len(out) < n {
		select {
		case e := <-events:
			out = append(out, string(e.Data))
		case <-deadline:
			t.Fatalf("timed out waiting for %d SSE events, got %d", n, len(out))
		}
	}
	return out
}

func TestWeatherPublishSnapshot(t *testing.T) {
	s, client := newSSEServer(t)
	w := &WeatherService{
		sse:     s,
		weather: []Day{{Name: "Sun 23 Aug"}},
		hourly:  []Hour{{Time: "12:00"}},
	}

	w.publishSnapshot()
	got := collectSSEEvents(t, client, "weather", 2)
	if len(got) != 2 {
		t.Fatalf("publishSnapshot() produced %d events, want 2 (current+hourly)", len(got))
	}
}

func TestWeatherPublishSnapshotEmptyWeather(t *testing.T) {
	s, client := newSSEServer(t)
	w := &WeatherService{sse: s}

	events := make(chan string, 1)
	go func() {
		_ = client.Subscribe("weather", func(msg *sse.Event) {
			if len(msg.Data) > 0 {
				events <- string(msg.Data)
			}
		})
	}()
	time.Sleep(300 * time.Millisecond) // let subscription register
	w.publishSnapshot()

	select {
	case e := <-events:
		t.Fatalf("publishSnapshot() with empty weather published unexpected event: %s", e)
	case <-time.After(300 * time.Millisecond):
		// expected: no events
	}
}

func TestWeatherServiceFetchAndPublish(t *testing.T) {
	setFetchWeather(t, func(meteo.Options) (meteo.WeatherResponse, error) {
		return testWeatherResponse(), nil
	})

	s, _ := newSSEServer(t)
	w := &WeatherService{sse: s, loc: time.UTC}

	if err := w.fetchAndPublish(); err != nil {
		t.Fatalf("fetchAndPublish() error = %v", err)
	}

	weather := w.GetCurrentWeather()
	if len(weather) != 2 {
		t.Fatalf("GetCurrentWeather() = %d days, want 2", len(weather))
	}
	day := weather[0]
	if day.Name == "" {
		t.Error("day.Name is empty")
	}
	if day.TemperatureMax != "25.0 °C" {
		t.Errorf("day.TemperatureMax = %q, want %q", day.TemperatureMax, "25.0 °C")
	}
	if day.More.CurrentTemperature != "21.5 °C" {
		t.Errorf("day.More.CurrentTemperature = %q, want %q", day.More.CurrentTemperature, "21.5 °C")
	}
	if w.GetCurrentHourly() == nil {
		t.Error("hourly not set after fetch")
	}
	if w.lastResponse == nil {
		t.Error("lastResponse not set after fetch")
	}
}

func TestWeatherFetchAndPublishError(t *testing.T) {
	setFetchWeather(t, func(meteo.Options) (meteo.WeatherResponse, error) {
		return meteo.WeatherResponse{}, errTest
	})

	w := &WeatherService{}
	if err := w.fetchAndPublish(); err == nil {
		t.Fatal("fetchAndPublish() expected error, got nil")
	}
}

func TestWeatherFetchIncompleteData(t *testing.T) {
	setFetchWeather(t, func(meteo.Options) (meteo.WeatherResponse, error) {
		return meteo.WeatherResponse{}, nil // empty daily/sunrise
	})

	w := &WeatherService{}
	if err := w.fetchAndPublish(); err == nil {
		t.Fatal("fetchAndPublish() expected error for incomplete data, got nil")
	}
}

func TestWeatherHasResponseChanged(t *testing.T) {
	w := &WeatherService{}
	base := testWeatherResponse()

	if !w.hasResponseChanged(&base) {
		t.Error("hasResponseChanged() = false when lastResponse is nil")
	}

	w.lastResponse = &base
	if w.hasResponseChanged(&base) {
		t.Error("hasResponseChanged() = true for identical response")
	}

	changed := testWeatherResponse()
	changed.Current.Temperature2m = 30.0
	if !w.hasResponseChanged(&changed) {
		t.Error("hasResponseChanged() = false when current temp changed")
	}

	changed = testWeatherResponse()
	changed.Daily.TemperatureMax[1] = 99.0
	if !w.hasResponseChanged(&changed) {
		t.Error("hasResponseChanged() = false when second day max changed")
	}
}

func TestWeatherRecomputeHourly(t *testing.T) {
	res := testWeatherResponse()
	s, _ := newSSEServer(t)
	w := &WeatherService{sse: s, loc: time.UTC, lastResponse: &res}

	w.recomputeHourly()
	if len(w.hourly) == 0 {
		t.Fatal("recomputeHourly() produced no hourly entries")
	}
}
