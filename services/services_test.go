package services

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/flohoss/godash/pkg/meteo"
)

func TestGetIcon(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		isDay bool
		want  string
	}{
		{"clear day", 0, true, "icon-[bi--sun-fill]"},
		{"clear night", 0, false, "icon-[bi--moon-fill]"},
		{"partly cloudy day", 1, true, "icon-[bi--cloud-sun-fill]"},
		{"partly cloudy night", 2, false, "icon-[bi--cloud-moon-fill]"},
		{"overcast", 3, true, "icon-[bi--cloud-fill]"},
		{"fog", 45, true, "icon-[bi--cloud-fog2-fill]"},
		{"drizzle", 51, true, "icon-[bi--cloud-drizzle-fill]"},
		{"rain heavy", 63, true, "icon-[bi--cloud-rain-heavy-fill]"},
		{"snow", 75, true, "icon-[bi--cloud-snow-fill]"},
		{"thunderstorm", 95, true, "icon-[bi--cloud-lightning-rain-fill]"},
		{"unknown code", 999, true, "icon-[bi--cloud-fill]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIcon(tt.code, tt.isDay); got != tt.want {
				t.Errorf("getIcon(%d, %v) = %q, want %q", tt.code, tt.isDay, got, tt.want)
			}
		})
	}
}

func TestHoursEqual(t *testing.T) {
	hour := Hour{Time: "12:00", Temperature: "21 °C"}
	if !slices.Equal([]Hour{hour}, []Hour{hour}) {
		t.Error("slices.Equal() = false for identical slices")
	}
	if slices.Equal([]Hour{hour}, []Hour{}) {
		t.Error("slices.Equal() = true for different lengths")
	}
	if slices.Equal([]Hour{hour}, []Hour{{Time: "13:00", Temperature: "21 °C"}}) {
		t.Error("slices.Equal() = true for different content")
	}
}

func TestNearestHourIndex(t *testing.T) {
	parsed := []time.Time{
		time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 23, 14, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name   string
		target time.Time
		want   int
	}{
		{"exact match", time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC), 1},
		{"before all", time.Date(2026, 8, 23, 11, 0, 0, 0, time.UTC), 0},
		{"after all", time.Date(2026, 8, 23, 15, 0, 0, 0, time.UTC), 2},
		{"between rounds to earlier", time.Date(2026, 8, 23, 13, 59, 0, 0, time.UTC), 2},
		{"equal distance ties to earlier", time.Date(2026, 8, 23, 13, 30, 0, 0, time.UTC), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nearestHourIndex(parsed, tt.target); got != tt.want {
				t.Errorf("nearestHourIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNearestHourIndexEmpty(t *testing.T) {
	if got := nearestHourIndex(nil, time.Now()); got != -1 {
		t.Errorf("nearestHourIndex(empty) = %d, want -1", got)
	}
}

func TestBuildHourlyEmptyTimes(t *testing.T) {
	res := &meteo.WeatherResponse{}
	hours := buildHourly(res, time.UTC)
	if len(hours) != 0 {
		t.Errorf("buildHourly() with empty times = %d hours, want 0", len(hours))
	}
}

func TestBuildHourlyInvalidTimeReturnsEmpty(t *testing.T) {
	var res meteo.WeatherResponse
	hourlyJSON := `{
		"hourly_units": {"temperature_2m": "°C", "wind_speed_10m": "km/h", "precipitation_probability": "%"},
		"hourly": {"time": ["not-a-timestamp"], "temperature_2m": [20.0]}
	}`
	if err := json.Unmarshal([]byte(hourlyJSON), &res); err != nil {
		t.Fatalf("failed to build response: %v", err)
	}
	hours := buildHourly(&res, time.UTC)
	if len(hours) != 0 {
		t.Errorf("buildHourly() with invalid timestamps = %d hours, want 0", len(hours))
	}
}

func TestBuildHourlyProducesUpToEight(t *testing.T) {
	loc := time.UTC
	start := time.Date(2026, 8, 23, 0, 0, 0, 0, loc)

	var timeStrs, tempStrs, codeStrs, isDayStrs, windStrs, precipStrs []string
	for i := 0; i < 24; i++ {
		timeStrs = append(timeStrs, start.Add(time.Duration(i)*time.Hour).Format("2006-01-02T15:04"))
		tempStrs = append(tempStrs, "20.0")
		codeStrs = append(codeStrs, "0")
		isDayStrs = append(isDayStrs, "1")
		windStrs = append(windStrs, "10")
		precipStrs = append(precipStrs, "10")
	}
	payload := `{
		"hourly_units": {"temperature_2m": "°C", "wind_speed_10m": "km/h", "precipitation_probability": "%"},
		"hourly": {
			"time": ["` + strings.Join(timeStrs, `","`) + `"],
			"temperature_2m": [` + strings.Join(tempStrs, ",") + `],
			"weather_code": [` + strings.Join(codeStrs, ",") + `],
			"is_day": [` + strings.Join(isDayStrs, ",") + `],
			"wind_speed_10m": [` + strings.Join(windStrs, ",") + `],
			"precipitation_probability": [` + strings.Join(precipStrs, ",") + `]
		}
	}`
	var res meteo.WeatherResponse
	if err := json.Unmarshal([]byte(payload), &res); err != nil {
		t.Fatalf("failed to build response: %v", err)
	}

	hours := buildHourly(&res, loc)
	if len(hours) != 8 {
		t.Errorf("buildHourly() = %d hours, want 8", len(hours))
	}
}

func TestRegisterAndPublishSnapshot(t *testing.T) {
	called := 0
	RegisterSnapshot("test-stream", func() { called++ })

	PublishSnapshot("test-stream")
	PublishSnapshot("test-stream")
	if called != 2 {
		t.Errorf("PublishSnapshot() called fn %d times, want 2", called)
	}

	PublishSnapshot("unknown-stream")
	if called != 2 {
		t.Errorf("PublishSnapshot(unknown) should not call fn, got %d calls", called)
	}

	RegisterSnapshot("test-stream", func() { called += 10 })
	PublishSnapshot("test-stream")
	if called != 12 {
		t.Errorf("PublishSnapshot() after re-register = %d, want 12", called)
	}
}
