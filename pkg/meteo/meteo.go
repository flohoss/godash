package meteo

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type WeatherResponse struct {
	CurrentUnits struct {
		Temperature2m       string `json:"temperature_2m"`
		ApparentTemperature string `json:"apparent_temperature"`
		RelativeHumidity    string `json:"relative_humidity_2m"`
		WindSpeed10m        string `json:"wind_speed_10m"`
	} `json:"current_units"`
	Current struct {
		Temperature2m       float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		RelativeHumidity    int     `json:"relative_humidity_2m"`
		WeatherCode         int     `json:"weather_code"`
		IsDay               float64 `json:"is_day"`
		WindSpeed10m        float64 `json:"wind_speed_10m"`
	} `json:"current"`
	HourlyUnits struct {
		Temperature2m            string `json:"temperature_2m"`
		WindSpeed10m             string `json:"wind_speed_10m"`
		PrecipitationProbability string `json:"precipitation_probability"`
	} `json:"hourly_units"`
	Hourly struct {
		Time                     []string  `json:"time"`
		Temperature2m            []float64 `json:"temperature_2m"`
		WeatherCode              []int     `json:"weather_code"`
		IsDay                    []int     `json:"is_day"`
		WindSpeed10m             []float64 `json:"wind_speed_10m"`
		PrecipitationProbability []int     `json:"precipitation_probability"`
	} `json:"hourly"`
	DailyUnits struct {
		TemperatureMax string `json:"temperature_2m_max"`
		TemperatureMin string `json:"temperature_2m_min"`
		Sunrise        string `json:"sunrise"`
		Sunset         string `json:"sunset"`
	} `json:"daily_units"`
	Daily struct {
		Time           []string  `json:"time"`
		WeatherCode    []int     `json:"weather_code"`
		TemperatureMax []float64 `json:"temperature_2m_max"`
		TemperatureMin []float64 `json:"temperature_2m_min"`
		Sunrise        []string  `json:"sunrise"`
		Sunset         []string  `json:"sunset"`
	} `json:"daily"`
}

type Options struct {
	Latitude  float64
	Longitude float64
	TimeZone  string
	Units     string
}

func GetWeather(options Options) (WeatherResponse, error) {
	baseURL := "https://api.open-meteo.com/v1/forecast"
	current := "temperature_2m,apparent_temperature,relative_humidity_2m,weather_code,is_day,wind_speed_10m"
	daily := "temperature_2m_max,temperature_2m_min,weather_code,sunrise,sunset"
	hourly := "temperature_2m,weather_code,is_day,wind_speed_10m,precipitation_probability"
	windUnit := "kmh"
	if options.Units == "fahrenheit" {
		windUnit = "mph"
	}
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", options.Latitude))
	params.Set("longitude", fmt.Sprintf("%f", options.Longitude))
	params.Set("timezone", options.TimeZone)
	params.Set("temperature_unit", options.Units)
	params.Set("wind_speed_unit", windUnit)
	params.Set("daily", daily)
	params.Set("current", current)
	params.Set("hourly", hourly)
	params.Set("forecast_days", "2")

	resp, err := httpClient.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return WeatherResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherResponse{}, fmt.Errorf("received non-OK HTTP status %d", resp.StatusCode)
	}

	var weatherData WeatherResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&weatherData); err != nil {
		slog.Error("failed to decode weather response", "error", err)
		return WeatherResponse{}, err
	}

	return weatherData, nil
}
