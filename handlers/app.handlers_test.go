package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flohoss/godash/services"
	"github.com/labstack/echo/v5"
)

type fakeSystemService struct {
	buffer services.Buffer
	static services.Static
}

func (f *fakeSystemService) GetBuffer() services.Buffer { return f.buffer }
func (f *fakeSystemService) GetStatic() services.Static { return f.static }

type fakeWeatherService struct {
	days  []services.Day
	hours []services.Hour
}

func (f *fakeWeatherService) GetCurrentWeather() []services.Day { return f.days }
func (f *fakeWeatherService) GetCurrentHourly() []services.Hour { return f.hours }

func TestHandleIndex(t *testing.T) {
	e := echo.New()
	app := NewAppHandler(
		&fakeSystemService{
			buffer: services.Buffer{
				CPU:  services.Detail{Value: "Intel", Percentage: 10},
				RAM:  services.Detail{Value: "8 GB", Percentage: 50},
				Disk: services.Detail{Value: "100 GB", Percentage: 40},
			},
			static: services.Static{CPU: "8 threads", RAM: "no swap", Disk: "1 partitions"},
		},
		&fakeWeatherService{
			days:  []services.Day{{Name: "Sun 23 Aug", TemperatureMax: "25 °C"}},
			hours: []services.Hour{{Time: "12:00", Temperature: "21 °C"}},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := app.handleIndex(c); err != nil {
		t.Fatalf("handleIndex() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.Len() == 0 {
		t.Error("response body is empty")
	}
}
