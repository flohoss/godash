package handlers

import (
	"bytes"
	"net/http"
	"os"

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

func shortCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		return next(c)
	}
}

func staticHandler(root string) echo.HandlerFunc {
	return func(c echo.Context) error {
		p := c.Param("*")
		full := root + "/" + p
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			return c.NoContent(http.StatusNotFound)
		}
		return c.File(full)
	}
}

func render(c echo.Context, cmp templ.Component) error {
	var buf bytes.Buffer
	if err := cmp.Render(c.Request().Context(), &buf); err != nil {
		return c.String(http.StatusInternalServerError, "render error")
	}
	c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	c.Response().Writer.WriteHeader(http.StatusOK)
	_, _ = c.Response().Writer.Write(buf.Bytes())
	return nil
}

func SetupRoutes(e *echo.Echo, sse *sse.Server, appHandler *AppHandler) {
	e.GET("/sse", echo.WrapHandler(sse))

	assets := e.Group("/assets", longCacheLifetime)
	assets.GET("/*", staticHandler("assets"))

	icons := e.Group("/icons", shortCacheLifetime)
	icons.GET("/*", staticHandler("config/icons"))

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	e.GET("/", appHandler.handleIndex)

	e.GET("/*", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/")
	})
}
