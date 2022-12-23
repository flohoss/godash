package main

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"strings"
	"time"
)

func (g *goDash) authMiddleware() echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		Skipper: func(c echo.Context) bool {
			return g.config.Password == "" || c.Path() == "/auth/login" || strings.Contains(c.Path(), "/static")
		},
		SigningKey:  []byte("secret"),
		TokenLookup: "cookie:" + g.cookieName(),
		AuthScheme:  "",
		ErrorHandlerWithContext: func(err error, c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
		},
	})
}

func (g *goDash) cookieName() string {
	return g.config.Title + "-auth"
}

func (g *goDash) setupCookie(c echo.Context, value string, expires time.Time) {
	c.SetCookie(&http.Cookie{
		Name:     g.cookieName(),
		Value:    value,
		Path:     "/",
		Domain:   g.config.Domain.Host,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

type jwtCustomClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

func (g *goDash) loginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login", map[string]interface{}{
		"Title": g.config.Title + " - Login",
	})
}

func (g *goDash) loginHandler(c echo.Context) error {
	name := c.FormValue("name")
	password := c.FormValue("password")

	if password != g.config.Password {
		return echo.ErrUnauthorized
	}
	expires := time.Now().Add(time.Hour * 72)
	claims := &jwtCustomClaims{name, jwt.StandardClaims{ExpiresAt: expires.Unix()}}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	g.setupCookie(c, t, expires)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (g *goDash) logoutHandler(c echo.Context) error {
	g.setupCookie(c, "", time.Now())
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
