package meteo

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func setHttpClient(t *testing.T, c *http.Client) {
	t.Helper()
	old := httpClient
	httpClient = c
	t.Cleanup(func() { httpClient = old })
}

func setBaseURL(t *testing.T, u string) {
	t.Helper()
	old := baseURL
	baseURL = u
	t.Cleanup(func() { baseURL = old })
}

func sampleWeatherJSON() string {
	return `{
		"current_units": {"temperature_2m": "°C", "wind_speed_10m": "km/h"},
		"current": {"temperature_2m": 21.5, "relative_humidity_2m": 60, "weather_code": 0, "is_day": 1, "wind_speed_10m": 12},
		"hourly_units": {"temperature_2m": "°C", "wind_speed_10m": "km/h", "precipitation_probability": "%"},
		"hourly": {"time": ["2026-08-23T12:00"], "temperature_2m": [21.0], "weather_code": [0], "is_day": [1], "wind_speed_10m": [12], "precipitation_probability": [20]},
		"daily_units": {"temperature_2m_max": "°C", "temperature_2m_min": "°C"},
		"daily": {"time": ["2026-08-23"], "weather_code": [0], "temperature_2m_max": [25.0], "temperature_2m_min": [18.0], "sunrise": ["2026-08-23T06:00"], "sunset": ["2026-08-23T20:00"]}
	}`
}

func TestGetWeatherBuildsQuery(t *testing.T) {
	var gotParams url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotParams = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(sampleWeatherJSON()))
	}))
	defer srv.Close()
	setHttpClient(t, srv.Client())
	setBaseURL(t, srv.URL)

	_, err := GetWeather(Options{
		Latitude:  52.5163,
		Longitude: 13.3776,
		TimeZone:  "Europe/Berlin",
		Units:     "celsius",
	})
	if err != nil {
		t.Fatalf("GetWeather() error = %v", err)
	}

	if gotParams.Get("latitude") != "52.516300" {
		t.Errorf("latitude = %q, want %q", gotParams.Get("latitude"), "52.516300")
	}
	if gotParams.Get("longitude") != "13.377600" {
		t.Errorf("longitude = %q, want %q", gotParams.Get("longitude"), "13.377600")
	}
	if gotParams.Get("timezone") != "Europe/Berlin" {
		t.Errorf("timezone = %q, want %q", gotParams.Get("timezone"), "Europe/Berlin")
	}
	if gotParams.Get("temperature_unit") != "celsius" {
		t.Errorf("temperature_unit = %q, want %q", gotParams.Get("temperature_unit"), "celsius")
	}
	if gotParams.Get("wind_speed_unit") != "kmh" {
		t.Errorf("wind_speed_unit = %q, want %q", gotParams.Get("wind_speed_unit"), "kmh")
	}
	if gotParams.Get("forecast_days") != "2" {
		t.Errorf("forecast_days = %q, want %q", gotParams.Get("forecast_days"), "2")
	}
}

func TestGetWeatherFahrenheitWindUnit(t *testing.T) {
	var gotWindUnit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotWindUnit = r.URL.Query().Get("wind_speed_unit")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(sampleWeatherJSON()))
	}))
	defer srv.Close()
	setHttpClient(t, srv.Client())
	setBaseURL(t, srv.URL)

	if _, err := GetWeather(Options{Units: "fahrenheit"}); err != nil {
		t.Fatalf("GetWeather() error = %v", err)
	}
	if gotWindUnit != "mph" {
		t.Errorf("wind_speed_unit = %q, want %q", gotWindUnit, "mph")
	}
}

func TestGetWeatherDecodesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(sampleWeatherJSON()))
	}))
	defer srv.Close()
	setHttpClient(t, srv.Client())
	setBaseURL(t, srv.URL)

	got, err := GetWeather(Options{})
	if err != nil {
		t.Fatalf("GetWeather() error = %v", err)
	}
	if got.Current.Temperature2m != 21.5 {
		t.Errorf("Temperature2m = %v, want 21.5", got.Current.Temperature2m)
	}
	if got.CurrentUnits.Temperature2m != "°C" {
		t.Errorf("units Temperature2m = %q, want °C", got.CurrentUnits.Temperature2m)
	}
	if len(got.Daily.Time) != 1 || got.Daily.Time[0] != "2026-08-23" {
		t.Errorf("Daily.Time = %v, want [2026-08-23]", got.Daily.Time)
	}
}

func TestGetWeatherNonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	setHttpClient(t, srv.Client())
	setBaseURL(t, srv.URL)

	_, err := GetWeather(Options{})
	if err == nil {
		t.Fatal("GetWeather() expected error for non-OK status, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status 500 in error, got: %v", err)
	}
}

func TestGetWeatherInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer srv.Close()
	setHttpClient(t, srv.Client())
	setBaseURL(t, srv.URL)

	_, err := GetWeather(Options{})
	if err == nil {
		t.Fatal("GetWeather() expected error for invalid JSON, got nil")
	}
}

func TestGetWeatherNetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	setBaseURL(t, srv.URL)
	srv.Close()
	setHttpClient(t, srv.Client())

	_, err := GetWeather(Options{})
	if err == nil {
		t.Fatal("GetWeather() expected error for network failure, got nil")
	}
}
