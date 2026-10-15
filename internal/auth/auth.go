package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"

	"github.com/flohoss/godash/config"
)

const (
	SessionCookieName = "godash_session"
	stateCookieName   = "godash_oidc_state"

	loginPath    = "/auth/login"
	callbackPath = "/auth/callback"

	stateCookieTTL = 10 * time.Minute
	stateSeparator = "|"

	stateTokenSize = 24

	defaultSessionTTL = 24 * time.Hour
	discoveryTimeout  = 10 * time.Second
)

var (
	ErrUnauthenticated     = errors.New("not authenticated")
	ErrProviderUnavailable = errors.New("identity provider is not reachable")

	scopes = []string{oidc.ScopeOpenID}
)

type User struct {
	Subject string
}

type providerState struct {
	key         string
	verifier    *oidc.IDTokenVerifier
	oauthConfig oauth2.Config
}

type Auth struct {
	providerMu sync.Mutex
	provider   *providerState
	wg         sync.WaitGroup
}

func New() *Auth {
	service := &Auth{}
	service.warmProvider()
	return service
}

func (a *Auth) Shutdown() {
	if a == nil {
		return
	}
	a.wg.Wait()
}

func (a *Auth) warmProvider() {
	if !a.Enabled() {
		return
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		_, _ = a.resolve()
	}()
}

func (a *Auth) Enabled() bool {
	return a != nil && a.settings().Enabled
}

func (a *Auth) settings() config.OIDCSettings {
	return config.GetAuth().OIDC
}

func (a *Auth) resolve() (*providerState, error) {
	settings := a.settings()
	key := providerKey(settings)

	a.providerMu.Lock()
	defer a.providerMu.Unlock()

	if a.provider != nil && a.provider.key == key {
		return a.provider, nil
	}

	provider, err := discoverProvider(settings)
	if err != nil {
		return nil, err
	}

	a.provider = provider
	return provider, nil
}

func providerKey(settings config.OIDCSettings) string {
	return settings.IssuerURL + "|" + settings.ClientID
}

func discoverProvider(settings config.OIDCSettings) (*providerState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, settings.IssuerURL)
	if err != nil {
		slog.Warn("OIDC discovery failed, will retry on the next request", "issuer", settings.IssuerURL, "error", err)
		return nil, fmt.Errorf("%w: discovery for %q failed: %w", ErrProviderUnavailable, settings.IssuerURL, err)
	}

	slog.Info("OIDC provider discovered", "issuer", settings.IssuerURL)

	return &providerState{
		key:         providerKey(settings),
		verifier:    provider.Verifier(&oidc.Config{ClientID: settings.ClientID}),
		oauthConfig: oauth2.Config{ClientID: settings.ClientID, Endpoint: provider.Endpoint(), Scopes: scopes},
	}, nil
}

func (a *Auth) StartLogin(c *echo.Context) error {
	provider, err := a.resolve()
	if err != nil {
		return err
	}

	state, err := randomToken(stateTokenSize)
	if err != nil {
		return err
	}

	verifier := oauth2.GenerateVerifier()

	a.clearLoginState(c)
	c.SetCookie(a.cookie(stateCookieName, state+stateSeparator+verifier, int(stateCookieTTL.Seconds()), time.Time{}))

	login := a.loginConfig(provider, c)
	c.Redirect(http.StatusFound, login.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)))
	return nil
}

func (a *Auth) ConsumeLoginState(c *echo.Context) (string, string, error) {
	cookie, err := c.Cookie(stateCookieName)
	a.clearLoginState(c)
	if err != nil {
		return "", "", ErrUnauthenticated
	}

	state, verifier, found := strings.Cut(cookie.Value, stateSeparator)
	if !found || state == "" || verifier == "" {
		return "", "", ErrUnauthenticated
	}

	return state, verifier, nil
}

func (a *Auth) Exchange(c *echo.Context, code, verifier string) (string, error) {
	provider, err := a.resolve()
	if err != nil {
		return "", err
	}

	login := a.loginConfig(provider, c)
	token, err := login.Exchange(c.Request().Context(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		return "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", errors.New("OIDC provider did not return an id_token")
	}

	idToken, err := provider.verifier.Verify(c.Request().Context(), rawIDToken)
	if err != nil {
		return "", fmt.Errorf("failed to verify id_token: %w", err)
	}

	if idToken.Subject == "" {
		return "", errors.New("OIDC provider returned an id_token without a subject")
	}

	return idToken.Subject, nil
}

func (a *Auth) loginConfig(provider *providerState, c *echo.Context) oauth2.Config {
	login := provider.oauthConfig
	login.ClientSecret = a.settings().ClientSecret
	login.RedirectURL = a.callbackURL(c)
	return login
}

func (a *Auth) callbackURL(c *echo.Context) string {
	return c.Scheme() + "://" + c.Request().Host + callbackPath
}

func (a *Auth) StartSession(c *echo.Context, subject string) {
	c.SetCookie(a.SessionCookie(subject))
}

func (a *Auth) SessionCookie(subject string) *http.Cookie {
	expiresAt := time.Now().Add(a.SessionTTL())
	value := signSession(a.settings().SessionKey, subject, expiresAt)
	return a.cookie(SessionCookieName, value, 0, expiresAt)
}

func (a *Auth) Authenticate(c *echo.Context) (*User, error) {
	cookie, err := c.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil, ErrUnauthenticated
	}

	subject, expiresAt, err := verifySession(a.settings().SessionKey, cookie.Value, time.Now())
	if err != nil {
		return nil, ErrUnauthenticated
	}

	a.refreshSession(c, subject, expiresAt)

	return &User{Subject: subject}, nil
}

func (a *Auth) refreshSession(c *echo.Context, subject string, expiresAt time.Time) {
	if time.Until(expiresAt) >= a.SessionTTL()/2 {
		return
	}

	renewed := time.Now().Add(a.SessionTTL())
	value := signSession(a.settings().SessionKey, subject, renewed)
	c.SetCookie(a.cookie(SessionCookieName, value, 0, renewed))
}

func (a *Auth) SessionTTL() time.Duration {
	ttl := a.settings().SessionTTL
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}
	return ttl
}

func (a *Auth) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !a.Enabled() {
				return next(c)
			}

			if _, err := a.Authenticate(c); err != nil {
				return c.Redirect(http.StatusFound, loginPath)
			}

			return next(c)
		}
	}
}

func (a *Auth) cookie(name, value string, maxAge int, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expires,
		HttpOnly: true,
		Secure:   a.settings().CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (a *Auth) clearLoginState(c *echo.Context) {
	c.SetCookie(a.cookie(stateCookieName, "", -1, time.Time{}))
}

func randomToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}
