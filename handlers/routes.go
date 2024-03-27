package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
)

func SetupRoutes(e *echo.Echo, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	if authHandler.env.SSODomain != "" {
		e.GET("/auth/", echo.WrapHandler(authHandler.authN))
	}

	secure := e.Group("/")
	if authHandler.env.SSODomain != "" {
		secure = e.Group("/", echo.WrapMiddleware(authHandler.middleware.RequireAuthentication()))
	}

	secure.GET("", appHandler.appHandler)
	secure.GET("sse", echo.WrapHandler(http.HandlerFunc(sse.ServeHTTP)))

	secure.Static("", "assets")
	secure.Static("storage/icons", "storage/icons")
}

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
