package handlers

import (
	"errors"
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
	startLogin        func(c *echo.Context) error
	consumeLoginState func(c *echo.Context) (string, string, error)
	exchange          func(c *echo.Context, code, verifier string) (string, error)
	startSession      func(c *echo.Context, subject string)
}

func (s *stubAuthService) StartLogin(c *echo.Context) error {
	if s.startLogin != nil {
		return s.startLogin(c)
	}
	return s.Auth.StartLogin(c)
}

func (s *stubAuthService) ConsumeLoginState(c *echo.Context) (string, string, error) {
	if s.consumeLoginState != nil {
		return s.consumeLoginState(c)
	}
	return s.Auth.ConsumeLoginState(c)
}

func (s *stubAuthService) Exchange(c *echo.Context, code, verifier string) (string, error) {
	if s.exchange != nil {
		return s.exchange(c, code, verifier)
	}
	return s.Auth.Exchange(c, code, verifier)
}

func (s *stubAuthService) StartSession(c *echo.Context, subject string) {
	if s.startSession != nil {
		s.startSession(c, subject)
		return
	}
	s.Auth.StartSession(c, subject)
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

// Everything except the OIDC handshake and the static assets sits behind the
// session check, so an unauthenticated request is redirected into the login flow.
func TestSetupRoutes_WithAuth_ProtectsTheDashboard(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	for _, path := range []string{"/", "/icons/favicon.ico", "/sse", "/anything"} {
		got, location := statusOf(e, http.MethodGet, path)
		if got != http.StatusFound || location != "/auth/login" {
			t.Errorf("%s = %d %q, want 302 to /auth/login", path, got, location)
		}
	}
}

// Static assets stay public so the login page can load its stylesheet and
// script without a session.
func TestSetupRoutes_WithAuth_ServesAssetsWithoutSession(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	got, location := statusOf(e, http.MethodGet, "/assets/css/style.css")
	if got == http.StatusFound && location == "/auth/login" {
		t.Error("/assets/css/style.css was redirected to the login page; assets must stay public")
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
	e, _ := setupRouterWithAuth(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(auth.SessionCookie("user-subject"))
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected the dashboard to be served with a valid session, got %d", rec.Code)
	}
}

// A stale or missing state cookie must not be treated as a successful login.
func TestSetupRoutes_WithAuth_CallbackRedirectsToLoginWithoutState(t *testing.T) {
	e, _ := setupRouterWithAuth(t)

	got, location := statusOf(e, http.MethodGet, "/auth/callback?code=x&state=y")
	if got != http.StatusFound || location != "/auth/login?error=expired" {
		t.Errorf("/auth/callback = %d %q, want 302 to /auth/login?error=expired", got, location)
	}
}

// A state mismatch means the callback does not belong to the login this browser
// started, so it must be discarded.
func TestSetupRoutes_WithAuth_CallbackRedirectsToLoginOnStateMismatch(t *testing.T) {
	loadOIDCConfig(t, true)

	service := &stubAuthService{Auth: auth.New()}
	service.consumeLoginState = func(*echo.Context) (string, string, error) {
		return "expected-state", "verifier", nil
	}

	appHandler := NewAppHandler(&fakeSystemService{}, sampleWeatherService())
	e := echo.New()
	SetupRoutes(e, nil, appHandler, NewAuthHandler(service))

	got, location := statusOf(e, http.MethodGet, "/auth/callback?code=x&state=different")
	if got != http.StatusFound || location != "/auth/login?error=expired" {
		t.Errorf("/auth/callback = %d %q, want 302 to /auth/login?error=expired", got, location)
	}
}

// A provider outage during the code exchange is reported as unavailable so the
// user is not told their credentials were wrong.
func TestSetupRoutes_WithAuth_CallbackReportsUnavailableProvider(t *testing.T) {
	loadOIDCConfig(t, true)

	service := &stubAuthService{Auth: auth.New()}
	service.consumeLoginState = func(*echo.Context) (string, string, error) {
		return "state", "verifier", nil
	}
	service.exchange = func(*echo.Context, string, string) (string, error) {
		return "", auth.ErrProviderUnavailable
	}

	appHandler := NewAppHandler(&fakeSystemService{}, sampleWeatherService())
	e := echo.New()
	SetupRoutes(e, nil, appHandler, NewAuthHandler(service))

	got, location := statusOf(e, http.MethodGet, "/auth/callback?code=x&state=state")
	if got != http.StatusFound || location != "/auth/login?error=unavailable" {
		t.Errorf("/auth/callback = %d %q, want 302 to /auth/login?error=unavailable", got, location)
	}
}

// A failed exchange sends the user back to the login page instead of a bare error.
func TestSetupRoutes_WithAuth_CallbackRedirectsToLoginOnFailedExchange(t *testing.T) {
	loadOIDCConfig(t, true)

	service := &stubAuthService{Auth: auth.New()}
	service.consumeLoginState = func(*echo.Context) (string, string, error) {
		return "state", "verifier", nil
	}
	service.exchange = func(*echo.Context, string, string) (string, error) {
		return "", errors.New("bad code")
	}

	appHandler := NewAppHandler(&fakeSystemService{}, sampleWeatherService())
	e := echo.New()
	SetupRoutes(e, nil, appHandler, NewAuthHandler(service))

	got, location := statusOf(e, http.MethodGet, "/auth/callback?code=x&state=state")
	if got != http.StatusFound || location != "/auth/login?error=failed" {
		t.Errorf("/auth/callback = %d %q, want 302 to /auth/login?error=failed", got, location)
	}
}

// A successful exchange starts the session and lands the user on the dashboard.
func TestSetupRoutes_WithAuth_CallbackStartsSessionOnSuccess(t *testing.T) {
	loadOIDCConfig(t, true)

	var subject string
	service := &stubAuthService{Auth: auth.New()}
	service.consumeLoginState = func(*echo.Context) (string, string, error) {
		return "state", "verifier", nil
	}
	service.exchange = func(*echo.Context, string, string) (string, error) {
		return "user-subject", nil
	}
	service.startSession = func(c *echo.Context, got string) {
		subject = got
		c.SetCookie(auth.SessionCookie(got))
	}

	appHandler := NewAppHandler(&fakeSystemService{}, sampleWeatherService())
	e := echo.New()
	SetupRoutes(e, nil, appHandler, NewAuthHandler(service))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/callback?code=x&state=state", nil))

	if rec.Code != http.StatusFound {
		t.Fatalf("/auth/callback = %d, want %d", rec.Code, http.StatusFound)
	}
	if location := rec.Header().Get(echo.HeaderLocation); location != "/" {
		t.Errorf("expected a redirect to /, got %q", location)
	}
	if subject != "user-subject" {
		t.Errorf("session subject = %q, want %q", subject, "user-subject")
	}
}

// With auth disabled the callback must not leak a handshake.
func TestSetupRoutes_WithoutAuth_CallbackIsNotFound(t *testing.T) {
	loadOIDCConfig(t, false)

	service := &stubAuthService{Auth: auth.New()}
	appHandler := NewAppHandler(&fakeSystemService{}, sampleWeatherService())

	e := echo.New()
	SetupRoutes(e, nil, appHandler, NewAuthHandler(service))

	if got, _ := statusOf(e, http.MethodGet, "/auth/callback?code=x&state=y"); got != http.StatusNotFound {
		t.Errorf("/auth/callback = %d, want %d", got, http.StatusNotFound)
	}
}
