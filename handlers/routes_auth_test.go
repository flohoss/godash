package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/internal/auth"
	"github.com/flohoss/godash/services"
)

func loadOIDCConfig(t *testing.T, enabled bool) {
	t.Helper()

	viper.Reset()
	viper.Set("time_zone", "UTC")
	viper.Set("auth.oidc.enabled", enabled)
	viper.Set("auth.oidc.session_key", "a-sufficiently-long-session-signing-key")
	if enabled {
		viper.Set("auth.oidc.issuer_url", "https://sso.example.com")
		viper.Set("auth.oidc.client_id", "godash")
		viper.Set("auth.oidc.client_secret", "test-secret")
	}

	if err := config.ValidateAndLoadConfig(); err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}
}

func setupRouterWithAuth(t *testing.T) (*echo.Echo, *stubAuthService) {
	t.Helper()

	loadOIDCConfig(t, true)

	service := &stubAuthService{Auth: auth.New()}

	appHandler := NewAppHandler(&fakeSystemService{}, sampleWeatherService())
	authHandler := NewAuthHandler(service)

	e := echo.New()
	SetupRoutes(e, nil, appHandler, authHandler)

	return e, service
}

func sampleWeatherService() *fakeWeatherService {
	return &fakeWeatherService{
		days:  []services.Day{{Name: "Sun 23 Aug", TemperatureMax: "25 °C"}},
		hours: []services.Hour{{Time: "12:00", Temperature: "21 °C"}},
	}
}

type stubAuthService struct {
	*auth.Auth
}

func statusOf(e *echo.Echo, method, target string) (int, string) {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec.Code, rec.Header().Get(echo.HeaderLocation)
}

// With auth disabled the OIDC routes must behave like any unknown route rather
// than leaking a handshake.
func TestSetupRoutes_WithoutAuth_DisablesTheOIDCHandshake(t *testing.T) {
	loadOIDCConfig(t, false)

	service := &stubAuthService{Auth: auth.New()}
	appHandler := NewAppHandler(&fakeSystemService{}, sampleWeatherService())
	authHandler := NewAuthHandler(service)

	e := echo.New()
	SetupRoutes(e, nil, appHandler, authHandler)

	for _, path := range []string{"/auth/login", "/auth/callback?code=x&state=y"} {
		if got, _ := statusOf(e, http.MethodGet, path); got != http.StatusNotFound {
			t.Errorf("%s = %d, want %d", path, got, http.StatusNotFound)
		}
	}
}

// Everything except the OIDC handshake sits behind the session check, so an
// unauthenticated request is redirected into the login flow.
func TestSetupRoutes_WithAuth_ProtectsTheDashboard(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	for _, path := range []string{"/", "/assets/css/style.css", "/icons/favicon.ico", "/sse", "/anything"} {
		got, location := statusOf(e, http.MethodGet, path)
		if got != http.StatusFound || location != "/auth/login" {
			t.Errorf("%s = %d %q, want 302 to /auth/login", path, got, location)
		}
	}
}

func TestSetupRoutes_WithAuth_KeepsHandshakeOpen(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	if got, _ := statusOf(e, http.MethodGet, "/auth/login"); got == http.StatusNotFound {
		t.Error("expected /auth/login to be reachable without a session")
	}
	if got, _ := statusOf(e, http.MethodGet, "/robots.txt"); got != http.StatusOK {
		t.Errorf("/robots.txt = %d, want %d", got, http.StatusOK)
	}
}

// A valid session lets the dashboard through without a redirect.
func TestSetupRoutes_WithAuth_ServesDashboardWithValidSession(t *testing.T) {
	e, service := setupRouterWithAuth(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(service.SessionCookie("user-subject"))
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected the dashboard to be served with a valid session, got %d", rec.Code)
	}
}
