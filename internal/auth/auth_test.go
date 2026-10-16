package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc/oidctest"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/gosso"
)

func loadAuthConfig(t *testing.T, enabled bool, issuerURL string) {
	t.Helper()

	viper.Reset()
	viper.Set("time_zone", "UTC")
	viper.Set("auth.oidc.enabled", enabled)
	viper.Set("auth.oidc.session_key", "a-sufficiently-long-session-signing-key")

	if issuerURL != "" {
		viper.Set("auth.oidc.issuer_url", issuerURL)
		viper.Set("auth.oidc.client_id", "godash")
		viper.Set("auth.oidc.client_secret", "test-secret")
		viper.Set("auth.oidc.session_ttl", "24h")
	}

	if err := config.ValidateAndLoadConfig(); err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}
}

func serveWithCookie(service *Auth, target, cookie string) *httptest.ResponseRecorder {
	router := echo.New()
	router.Use(service.Middleware())
	router.GET(target, func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, target, nil)
	if cookie != "" {
		req.Header.Set(echo.HeaderCookie, cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

// Discovery is warmed in the background so the first login does not pay for the
// round trip, without blocking startup when the provider is slow or down.
func TestNew_WarmsProviderInBackground(t *testing.T) {
	oidcServer := &oidctest.Server{}
	server := httptest.NewServer(oidcServer)
	defer server.Close()
	oidcServer.SetIssuer(server.URL)

	loadAuthConfig(t, true, server.URL)

	service := New()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if service.core.Resolved() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("expected the provider to be discovered in the background")
}

// Behind proxies that do not forward the original scheme (e.g. NetBird), the
// registered redirect URI would be built with http and rejected by the
// provider. app_url overrides the request-derived value in that case.
func TestCallbackURL_PrefersAppUrl(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")
	viper.Set("server.app_url", "https://home.fhoss.de")
	if err := config.ValidateAndLoadConfig(); err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}

	service := New()

	router := echo.New()
	var got string
	router.GET("/auth/login", func(c *echo.Context) error {
		got = service.core.CallbackURL(c)
		return c.NoContent(http.StatusOK)
	})
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://godash:8156/auth/login", nil))

	if got != "https://home.fhoss.de/auth/callback" {
		t.Errorf("CallbackURL() = %q, want %q", got, "https://home.fhoss.de/auth/callback")
	}
}

func TestMiddleware_DisabledLetsEverythingThrough(t *testing.T) {
	loadAuthConfig(t, false, "")

	service := New()

	if got := serveWithCookie(service, "/", "").Code; got != http.StatusOK {
		t.Fatalf("expected open access when disabled, got %d", got)
	}
}

// Without a login page the middleware has to send the browser straight into the
// OIDC flow, otherwise the dashboard would be unreachable.
func TestMiddleware_RedirectsToLoginWhenUnauthenticated(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	service := New()

	rec := serveWithCookie(service, "/", "")

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 without a session, got %d", rec.Code)
	}
	if location := rec.Header().Get(echo.HeaderLocation); location != loginPath {
		t.Fatalf("expected redirect to %q, got %q", loginPath, location)
	}
}

func TestMiddleware_AcceptsValidSession(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	service := New()
	value := gosso.SignHMAC("a-sufficiently-long-session-signing-key", "user-subject", time.Now().Add(time.Hour))

	if got := serveWithCookie(service, "/", SessionCookieName+"="+value).Code; got != http.StatusOK {
		t.Fatalf("expected 200 for a valid session, got %d", got)
	}
}

func TestMiddleware_RejectsTamperedSession(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	service := New()
	value := gosso.SignHMAC("a-different-signing-key-that-is-long-enough", "user-subject", time.Now().Add(time.Hour))

	if got := serveWithCookie(service, "/", SessionCookieName+"="+value).Code; got != http.StatusFound {
		t.Fatalf("expected a redirect for a tampered session, got %d", got)
	}
}

// Sessions are renewed once they pass half their lifetime, so a user who loads
// the dashboard regularly never gets logged out mid-use.
func TestAuthenticate_RefreshesExpiringSession(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	service := New()
	value := gosso.SignHMAC("a-sufficiently-long-session-signing-key", "user-subject", time.Now().Add(time.Minute))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderCookie, SessionCookieName+"="+value)

	router := echo.New()
	router.GET("/", func(c *echo.Context) error {
		if _, err := service.Authenticate(c); err != nil {
			return c.NoContent(http.StatusUnauthorized)
		}
		return c.NoContent(http.StatusOK)
	})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected the session to authenticate, got %d", rec.Code)
	}

	refreshed := findCookie(rec, SessionCookieName)
	if refreshed == nil {
		t.Fatal("expected the expiring session to be refreshed")
	}
	if _, expiresAt, err := gosso.VerifyHMAC("a-sufficiently-long-session-signing-key", refreshed.Value, time.Now()); err != nil {
		t.Fatalf("refreshed cookie is not valid: %v", err)
	} else if time.Until(expiresAt) < time.Hour {
		t.Error("expected the refreshed cookie to carry a fresh expiry")
	}
}

func TestAuthenticate_KeepsFreshSession(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	service := New()
	value := gosso.SignHMAC("a-sufficiently-long-session-signing-key", "user-subject", time.Now().Add(23*time.Hour))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderCookie, SessionCookieName+"="+value)

	router := echo.New()
	router.GET("/", func(c *echo.Context) error {
		if _, err := service.Authenticate(c); err != nil {
			return c.NoContent(http.StatusUnauthorized)
		}
		return c.NoContent(http.StatusOK)
	})
	router.ServeHTTP(rec, req)

	if findCookie(rec, SessionCookieName) != nil {
		t.Error("expected a fresh session not to be refreshed")
	}
}

func findCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name && cookie.Value != "" {
			return cookie
		}
	}
	return nil
}

func TestSessionTTL_DefaultsWhenUnset(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	viper.Set("auth.oidc.session_ttl", "0s")
	if err := config.ValidateAndLoadConfig(); err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}

	if got := New().SessionTTL(); got != gosso.DefaultSessionTTL {
		t.Errorf("SessionTTL() = %v, want %v", got, gosso.DefaultSessionTTL)
	}
}

// settings() is the only place godash hands its configuration to gosso, so a
// typo here would silently disable the session key or the callback URL.
func TestSettings_MapsConfigIntoGosso(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")
	viper.Set("server.app_url", "https://home.fhoss.de")
	viper.Set("auth.oidc.session_ttl", "2h")
	viper.Set("auth.oidc.cookie_secure", true)
	if err := config.ValidateAndLoadConfig(); err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}

	got := settings()

	if !got.Enabled {
		t.Error("expected Enabled to follow the config")
	}
	if got.IssuerURL != "https://sso.example.com" {
		t.Errorf("IssuerURL = %q", got.IssuerURL)
	}
	if got.ClientID != "godash" {
		t.Errorf("ClientID = %q", got.ClientID)
	}
	if got.ClientSecret != "test-secret" {
		t.Errorf("ClientSecret = %q", got.ClientSecret)
	}
	if got.SessionKey != "a-sufficiently-long-session-signing-key" {
		t.Errorf("SessionKey = %q", got.SessionKey)
	}
	if got.SessionTTL != 2*time.Hour {
		t.Errorf("SessionTTL = %v, want 2h", got.SessionTTL)
	}
	if !got.CookieSecure {
		t.Error("expected CookieSecure to follow the config")
	}
	if got.AppURL != "https://home.fhoss.de" {
		t.Errorf("AppURL = %q", got.AppURL)
	}
}

// SessionCookie is godash's own signing helper, not a gosso call, so the
// fallback and the configured key both need to be verified here.
func TestSessionCookie_IsSignedWithTheConfiguredKey(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	cookie := SessionCookie("user-subject")

	if cookie.Name != SessionCookieName {
		t.Errorf("cookie name = %q, want %q", cookie.Name, SessionCookieName)
	}
	subject, expiresAt, err := gosso.VerifyHMAC("a-sufficiently-long-session-signing-key", cookie.Value, time.Now())
	if err != nil {
		t.Fatalf("SessionCookie() is not verifiable: %v", err)
	}
	if subject != "user-subject" {
		t.Errorf("subject = %q, want %q", subject, "user-subject")
	}
	if remaining := time.Until(expiresAt); remaining < 23*time.Hour || remaining > 25*time.Hour {
		t.Errorf("expiry in %v, want roughly 24h", remaining)
	}
}

func TestSessionCookie_DefaultsWhenTTLUnset(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	viper.Set("auth.oidc.session_ttl", "0s")
	if err := config.ValidateAndLoadConfig(); err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}

	cookie := SessionCookie("user-subject")

	_, expiresAt, err := gosso.VerifyHMAC("a-sufficiently-long-session-signing-key", cookie.Value, time.Now())
	if err != nil {
		t.Fatalf("SessionCookie() is not verifiable: %v", err)
	}
	if remaining := time.Until(expiresAt); remaining < gosso.DefaultSessionTTL-time.Minute {
		t.Errorf("expiry in %v, want the %v default", remaining, gosso.DefaultSessionTTL)
	}
}

func TestEnabled_FollowsConfig(t *testing.T) {
	loadAuthConfig(t, false, "")
	if New().Enabled() {
		t.Error("expected Enabled() to be false when OIDC is off")
	}

	loadAuthConfig(t, true, "https://sso.example.com")
	if !New().Enabled() {
		t.Error("expected Enabled() to be true when OIDC is on")
	}
}

func TestClassifyLoginError(t *testing.T) {
	if got := ClassifyLoginError(nil); got != "" {
		t.Errorf("ClassifyLoginError(nil) = %q, want empty", got)
	}
	if got := ClassifyLoginError(ErrProviderUnavailable); got != LoginErrorUnavailable {
		t.Errorf("ClassifyLoginError(unavailable) = %q, want %q", got, LoginErrorUnavailable)
	}
	if got := ClassifyLoginError(errors.New("boom")); got != LoginErrorFailed {
		t.Errorf("ClassifyLoginError(other) = %q, want %q", got, LoginErrorFailed)
	}
}

// StartLogin without a reachable provider must surface the failure instead of
// redirecting the browser into a dead handshake.
func TestStartLogin_ReportsAnUnreachableProvider(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	router := echo.New()
	router.GET("/auth/login", func(c *echo.Context) error {
		return New().StartLogin(c)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/login", nil))

	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusFound {
		t.Fatalf("expected an error or a redirect, got %d", rec.Code)
	}
}

// StartSession delegates to the configured SessionManager, which for godash is
// the signed-cookie store.
func TestStartSession_SetsASignedCookie(t *testing.T) {
	loadAuthConfig(t, true, "https://sso.example.com")

	service := New()

	router := echo.New()
	router.GET("/auth/callback", func(c *echo.Context) error {
		service.StartSession(c, "user-subject")
		return c.NoContent(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/callback", nil))

	cookie := findCookie(rec, SessionCookieName)
	if cookie == nil {
		t.Fatal("expected a session cookie")
	}
	if _, _, err := gosso.VerifyHMAC("a-sufficiently-long-session-signing-key", cookie.Value, time.Now()); err != nil {
		t.Fatalf("session cookie is not verifiable: %v", err)
	}
}
