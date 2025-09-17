package meteo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
	} `json:"current_units"`
	Current struct {
		Time                string  `json:"time"`
		Interval            int     `json:"interval"`
		Temperature2m       float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		RelativeHumidity    int     `json:"relative_humidity_2m"`
		WeatherCode         int     `json:"weather_code"`
		IsDay               int     `json:"is_day"`
	} `json:"current"`
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
	current := "temperature_2m,apparent_temperature,relative_humidity_2m,weather_code,is_day"
	daily := "temperature_2m_max,temperature_2m_min,weather_code,sunrise,sunset"
	params := fmt.Sprintf("?latitude=%f&longitude=%f&timezone=%s&temperature_unit=%s&daily=%s&current=%s", options.Latitude, options.Longitude, options.TimeZone, options.Units, daily, current)

	resp, err := http.Get(url + params)
	if err != nil {
		return WeatherResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherResponse{}, fmt.Errorf("received non-OK HTTP status %d", resp.StatusCode)
	}

	var weatherData WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		log.Fatal(err)
	}

	return weatherData, nil
}
