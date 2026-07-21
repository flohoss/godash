package meteo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type WeatherResponse struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationTimeMS     float64 `json:"generationtime_ms"`
	UTCOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
	CurrentUnits         struct {
		Time                string `json:"time"`
		Interval            string `json:"interval"`
		Temperature2m       string `json:"temperature_2m"`
		ApparentTemperature string `json:"apparent_temperature"`
		RelativeHumidity    string `json:"relative_humidity_2m"`
		WeatherCode         string `json:"weather_code"`
		IsDay               string `json:"is_day"`
		WindSpeed10m        string `json:"wind_speed_10m"`
		WindDirection10m    string `json:"wind_direction_10m"`
	} `json:"current_units"`
	Current struct {
		Time                string  `json:"time"`
		Interval            int     `json:"interval"`
		Temperature2m       float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		RelativeHumidity    int     `json:"relative_humidity_2m"`
		WeatherCode         int     `json:"weather_code"`
		IsDay               float64 `json:"is_day"`
		WindSpeed10m        float64 `json:"wind_speed_10m"`
		WindDirection10m    int     `json:"wind_direction_10m"`
	} `json:"current"`
	HourlyUnits struct {
		Time                     string `json:"time"`
		Temperature2m            string `json:"temperature_2m"`
		ApparentTemperature      string `json:"apparent_temperature"`
		WeatherCode              string `json:"weather_code"`
		IsDay                    string `json:"is_day"`
		WindSpeed10m             string `json:"wind_speed_10m"`
		WindDirection10m         string `json:"wind_direction_10m"`
		PrecipitationProbability string `json:"precipitation_probability"`
	} `json:"hourly_units"`
	Hourly struct {
		Time                     []string  `json:"time"`
		Temperature2m            []float64 `json:"temperature_2m"`
		ApparentTemperature      []float64 `json:"apparent_temperature"`
		WeatherCode              []int     `json:"weather_code"`
		IsDay                    []int     `json:"is_day"`
		WindSpeed10m             []float64 `json:"wind_speed_10m"`
		WindDirection10m         []int     `json:"wind_direction_10m"`
		PrecipitationProbability []int     `json:"precipitation_probability"`
	} `json:"hourly"`
	DailyUnits struct {
		Time           string `json:"time"`
		WeatherCode    string `json:"weather_code"`
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
	url := "https://api.open-meteo.com/v1/forecast"
	current := "temperature_2m,apparent_temperature,relative_humidity_2m,weather_code,is_day,wind_speed_10m,wind_direction_10m"
	daily := "temperature_2m_max,temperature_2m_min,weather_code,sunrise,sunset"
	hourly := "temperature_2m,apparent_temperature,weather_code,is_day,wind_speed_10m,wind_direction_10m,precipitation_probability"
	windUnit := "kmh"
	if options.Units == "fahrenheit" {
		windUnit = "mph"
	}
	params := fmt.Sprintf("?latitude=%f&longitude=%f&timezone=%s&temperature_unit=%s&wind_speed_unit=%s&daily=%s&current=%s&hourly=%s&forecast_days=2", options.Latitude, options.Longitude, options.TimeZone, options.Units, windUnit, daily, current, hourly)

	resp, err := httpClient.Get(url + params)
	if err != nil {
		return WeatherResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherResponse{}, fmt.Errorf("received non-OK HTTP status %d", resp.StatusCode)
	}

	var weatherData WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		log.Printf("failed to decode weather response: %v", err)
		return WeatherResponse{}, err
	}

	return weatherData, nil
}
