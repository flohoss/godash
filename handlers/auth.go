package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/godash/internal/auth"
	authview "github.com/flohoss/godash/views/auth"
	"github.com/flohoss/gosso/v2"
)

type AuthService interface {
	Enabled() bool
	StartLogin(c *echo.Context) error
	HandleCallback(c *echo.Context) (*gosso.LoginResult, *gosso.LoginFailure)
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

	result, failure := ah.Auth.HandleCallback(c)
	if failure != nil {
		gosso.LogFailure(failure)

		return redirectToLogin(c, failure.Code)
	}

	slog.Info("User logged in", "subject", result.Subject)

	return c.Redirect(http.StatusFound, result.Redirect)
}

func redirectToLogin(c *echo.Context, code auth.LoginError) error {
	return c.Redirect(http.StatusFound, "/auth/login?error="+string(code))
}

func loginErrorFromQuery(raw string) auth.LoginError {
	code := auth.LoginError(raw)
	if code.Valid() {
		return code
	}

	return ""
}

func isKnownProvider(provider string) bool {
	for _, option := range config.GetAuthProviders() {
		if option.Provider == provider {
			return true
		}
	}
	return false
}
