package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
)

func longCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000")
		return next(c)
	}
}

func render(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func SetupRoutes(e *echo.Echo, sse *sse.Server, appHandler *AppHandler) {
	e.GET("/sse", echo.WrapHandler(sse))

	assets := e.Group("/assets", longCacheLifetime)
	assets.Static("/", "assets")

	icons := e.Group("/icons", longCacheLifetime)
	icons.Static("/", "config/icons")

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	e.GET("/", appHandler.handleIndex)

	e.Any("/*", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/")
	})
}
