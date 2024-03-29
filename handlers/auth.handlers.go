package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/logto-io/go/client"
	"gitlab.unjx.de/flohoss/godash/internal/env"
)

func NewAuthHandler(env *env.Config) *AuthHandler {
	return &AuthHandler{
		env: env,
		logtoConfig: &client.LogtoConfig{
			Endpoint:  env.SSOEndpoint,
			AppId:     env.SSOAppId,
			AppSecret: env.SSOAppSecret,
			Resources: env.SSOResources,
		},
	}
}

type AuthHandler struct {
	env         *env.Config
	logtoConfig *client.LogtoConfig
}

func (authHandler *AuthHandler) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logtoClient := client.NewLogtoClient(
			authHandler.logtoConfig,
			NewSessionStorage(c),
		)
		if !logtoClient.IsAuthenticated() {
			return c.Redirect(http.StatusTemporaryRedirect, "/sign-in")
		}
		return next(c)
	}
}

func (authHandler *AuthHandler) signInHandler(c echo.Context) error {
	logtoClient := client.NewLogtoClient(
		authHandler.logtoConfig,
		NewSessionStorage(c),
	)
	signInUri, err := logtoClient.SignIn(authHandler.env.PublicUrl + "/sign-in-callback")
	if err != nil {
		slog.Error("cannot process sign in request", "err", err)
		return echo.ErrInternalServerError
	}
	return c.Redirect(http.StatusTemporaryRedirect, signInUri)
}

func (authHandler *AuthHandler) signInCallbackHandler(c echo.Context) error {
	logtoClient := client.NewLogtoClient(
		authHandler.logtoConfig,
		NewSessionStorage(c),
	)
	err := logtoClient.HandleSignInCallback(c.Request())
	if err != nil {
		slog.Error("cannot process sign in callback", "err", err)
		return echo.ErrInternalServerError
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (authHandler *AuthHandler) signOutCallbackHandler(c echo.Context) error {
	logtoClient := client.NewLogtoClient(
		authHandler.logtoConfig,
		NewSessionStorage(c),
	)
	signOutUri, err := logtoClient.SignOut(authHandler.env.PublicUrl)
	if err != nil {
		slog.Error("cannot process sign out", "err", err)
		return echo.ErrInternalServerError
	}
	return c.Redirect(http.StatusTemporaryRedirect, signOutUri)
}
