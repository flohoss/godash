package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/flohoss/godash/internal/auth"
)

type AuthService interface {
	Enabled() bool
	StartLogin(c *echo.Context) error
	ConsumeLoginState(c *echo.Context) (string, string, error)
	Exchange(c *echo.Context, code, verifier string) (string, error)
	StartSession(c *echo.Context, subject string)
	Authenticate(c *echo.Context) (*auth.User, error)
	Middleware() echo.MiddlewareFunc
}

type AuthHandler struct {
	Auth AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{Auth: service}
}

func (ah *AuthHandler) HandleStartLogin(c *echo.Context) error {
	if !ah.Auth.Enabled() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}

	if err := ah.Auth.StartLogin(c); err != nil {
		if errors.Is(err, auth.ErrProviderUnavailable) {
			slog.Error("OIDC provider unavailable", "error", err)
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Identity provider is not reachable")
		}

		slog.Error("Failed to start login", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start login")
	}

	return nil
}

func (ah *AuthHandler) HandleCallback(c *echo.Context) error {
	if !ah.Auth.Enabled() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}

	if providerError := c.QueryParam("error"); providerError != "" {
		slog.Warn("OIDC provider rejected login", "error", providerError)
		return echo.NewHTTPError(http.StatusUnauthorized, "The identity provider rejected the login")
	}

	state, verifier, err := ah.Auth.ConsumeLoginState(c)
	if err != nil {
		slog.Warn("OIDC login session expired or missing")
		return echo.NewHTTPError(http.StatusUnauthorized, "The login session expired, please try again")
	}

	if c.QueryParam("state") != state {
		slog.Warn("OIDC state mismatch, discarding callback")
		return echo.NewHTTPError(http.StatusUnauthorized, "The login session expired, please try again")
	}

	subject, err := ah.Auth.Exchange(c, c.QueryParam("code"), verifier)
	if err != nil {
		if errors.Is(err, auth.ErrProviderUnavailable) {
			slog.Error("OIDC provider unavailable", "error", err)
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Identity provider is not reachable")
		}

		slog.Error("OIDC login failed", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication failed")
	}

	ah.Auth.StartSession(c, subject)
	slog.Info("User logged in", "subject", subject)

	return c.Redirect(http.StatusFound, "/")
}
