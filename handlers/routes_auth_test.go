package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
	service.startLogin = func(c *echo.Context) error {
		return auth.ErrProviderUnavailable
	}

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
	startLogin func(c *echo.Context) error
}

func (s *stubAuthService) StartLogin(c *echo.Context) error {
	if s.startLogin != nil {
		return s.startLogin(c)
	}
	return s.Auth.StartLogin(c)
}

func statusOf(e *echo.Echo, method, target string) (int, string) {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec.Code, rec.Header().Get(echo.HeaderLocation)
}

func bodyOf(e *echo.Echo, method, target string) (int, string) {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec.Code, rec.Body.String()
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

// Unknown provider identifiers must not start a handshake.
func TestSetupRoutes_WithAuth_RejectsUnknownProvider(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	if got, _ := statusOf(e, http.MethodGet, "/auth/unknown"); got != http.StatusNotFound {
		t.Errorf("/auth/unknown = %d, want %d", got, http.StatusNotFound)
	}
}

// The login page always renders, even with a single provider, and offers a
// button that starts the handshake.
func TestSetupRoutes_WithAuth_LoginPageRendersProviderButtons(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	got, body := bodyOf(e, http.MethodGet, "/auth/login")
	if got != http.StatusOK {
		t.Fatalf("/auth/login = %d, want %d", got, http.StatusOK)
	}
	if !strings.Contains(body, "Sign in with OIDC") {
		t.Errorf("expected the OIDC button, got %q", body)
	}
	if !strings.Contains(body, "/auth/oidc") {
		t.Errorf("expected the OIDC button to link to /auth/oidc, got %q", body)
	}
}

// A provider that cannot be reached sends the user back to the login page with
// an error signal instead of bouncing them in a redirect loop.
func TestSetupRoutes_WithAuth_HandshakeFailureRedirectsToLogin(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	got, location := statusOf(e, http.MethodGet, "/auth/oidc")
	if got != http.StatusFound || location != "/auth/login?error=unavailable" {
		t.Errorf("/auth/oidc = %d %q, want 302 to /auth/login?error=unavailable", got, location)
	}
}

// A provider-side rejection sends the user back to the login page with an error
// signal instead of a bare JSON error.
func TestSetupRoutes_WithAuth_CallbackRedirectsToLoginOnProviderError(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	got, location := statusOf(e, http.MethodGet, "/auth/callback?error=access_denied")
	if got != http.StatusFound || location != "/auth/login?error=rejected" {
		t.Errorf("/auth/callback?error=access_denied = %d %q, want 302 to /auth/login?error=rejected", got, location)
	}
}

// The login page renders the provider buttons and maps the error code to a
// human-readable message.
func TestSetupRoutes_WithAuth_RendersLoginPageWithError(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	got, body := bodyOf(e, http.MethodGet, "/auth/login?error=expired")
	if got != http.StatusOK {
		t.Fatalf("/auth/login?error=expired = %d, want %d", got, http.StatusOK)
	}
	if !strings.Contains(body, "The login session expired") {
		t.Errorf("expected the expired-session message in the login page, got %q", body)
	}
	if !strings.Contains(body, "Sign in with OIDC") {
		t.Errorf("expected the OIDC button in the login page, got %q", body)
	}
	if !strings.Contains(body, "/auth/oidc") {
		t.Errorf("expected the OIDC button to link to /auth/oidc, got %q", body)
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
