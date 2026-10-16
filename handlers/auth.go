package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/internal/auth"
	authview "github.com/flohoss/godash/views/auth"
)

type AuthService interface {
	Enabled() bool
	StartLogin(c *echo.Context) error
	ConsumeLoginState(c *echo.Context) (string, string, error)
	Exchange(c *echo.Context, code, verifier string) (string, error)
	StartSession(c *echo.Context, subject string)
	Authenticate(c *echo.Context) (*auth.User, error)
	Middleware() echo.MiddlewareFunc
	Shutdown()
}

type AuthHandler struct {
	Auth AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{Auth: service}
}

func (ah *AuthHandler) HandleLoginPage(c *echo.Context) error {
	if !ah.Auth.Enabled() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}

	code := loginErrorFromQuery(c.QueryParam("error"))

	return render(c, authview.Login(config.GetTitle(), config.GetAuthProviders(), code))
}

func (ah *AuthHandler) HandleStartLogin(c *echo.Context) error {
	if !ah.Auth.Enabled() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}

	if !isKnownProvider(c.Param("provider")) {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}

	if err := ah.Auth.StartLogin(c); err != nil {
		slog.Error("Failed to start login", "error", err)
		return redirectToLogin(c, auth.ClassifyLoginError(err))
	}

	return nil
}

func (ah *AuthHandler) HandleCallback(c *echo.Context) error {
	if !ah.Auth.Enabled() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}

	if providerError := c.QueryParam("error"); providerError != "" {
		slog.Warn("OIDC provider rejected login", "error", providerError)
		return redirectToLogin(c, auth.LoginErrorRejected)
	}

	state, verifier, err := ah.Auth.ConsumeLoginState(c)
	if err != nil {
		slog.Warn("OIDC login session expired or missing")
		return redirectToLogin(c, auth.LoginErrorExpired)
	}

	if c.QueryParam("state") != state {
		slog.Warn("OIDC state mismatch, discarding callback")
		return redirectToLogin(c, auth.LoginErrorExpired)
	}

	subject, err := ah.Auth.Exchange(c, c.QueryParam("code"), verifier)
	if err != nil {
		if errors.Is(err, auth.ErrProviderUnavailable) {
			slog.Error("OIDC provider unavailable", "error", err)
		} else {
			slog.Error("OIDC login failed", "error", err)
		}
		return redirectToLogin(c, auth.ClassifyLoginError(err))
	}

	ah.Auth.StartSession(c, subject)
	slog.Info("User logged in", "subject", subject)

	return c.Redirect(http.StatusFound, "/")
}

func redirectToLogin(c *echo.Context, code auth.LoginError) error {
	return c.Redirect(http.StatusFound, "/auth/login?error="+string(code))
}

func loginErrorFromQuery(raw string) auth.LoginError {
	code := auth.LoginError(raw)
	switch code {
	case auth.LoginErrorRejected, auth.LoginErrorExpired, auth.LoginErrorUnavailable, auth.LoginErrorFailed:
		return code
	default:
		return ""
	}
}

func isKnownProvider(provider string) bool {
	for _, option := range config.GetAuthProviders() {
		if option.Provider == provider {
			return true
		}
	}
	return false
}
