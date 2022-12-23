package main

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

func (g *goDash) authMiddleware() echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/auth/login"
		},
		SigningKey:  []byte("secret"),
		TokenLookup: "cookie:" + g.cookieName(),
		AuthScheme:  "",
	})
}

type jwtCustomClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

func (g *goDash) cookieName() string {
	return g.config.Title + "-auth"
}

func (g *goDash) login(c echo.Context) error {
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

	c.SetCookie(&http.Cookie{
		Name:     g.cookieName(),
		Value:    t,
		Path:     "/",
		Domain:   g.config.Domain.Host,
		Expires:  expires,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
