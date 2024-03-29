package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
)

func SetupRoutes(e *echo.Echo, sse *sse.Server, appHandler *AppHandler, authHandler *AuthHandler) {
	if authHandler.env.LogtoEndpoint != "" {
		e.GET("/sign-in", authHandler.signInHandler)
		e.GET("/sign-in-callback", authHandler.signInCallbackHandler)
		e.GET("/sign-out", authHandler.signOutCallbackHandler)
	}

	secure := e.Group("/")
	if authHandler.env.LogtoEndpoint != "" {
		secure = e.Group("/", authHandler.logtoMiddleware)
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
