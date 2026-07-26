package handlers

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
	"github.com/r3labs/sse/v2"
)

func longCacheLifetime(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000")
		return next(c)
	}
}

func sseConnectionLimiter(maxPerIP int) echo.MiddlewareFunc {
	var mu sync.Mutex
	counts := map[string]int{}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ip := c.RealIP()
			mu.Lock()
			if counts[ip] >= maxPerIP {
				mu.Unlock()
				return c.NoContent(http.StatusServiceUnavailable)
			}
			counts[ip]++
			mu.Unlock()

			defer func() {
				mu.Lock()
				counts[ip]--
				if counts[ip] <= 0 {
					delete(counts, ip)
				}
				mu.Unlock()
			}()

			return next(c)
		}
	}
}

func render(c *echo.Context, cmp templ.Component) error {
	var buf bytes.Buffer
	if err := cmp.Render(c.Request().Context(), &buf); err != nil {
		return c.String(http.StatusInternalServerError, "render error")
	}
	w := c.Response()
	w.Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
	return nil
}

func SetupRoutes(e *echo.Echo, sse *sse.Server, appHandler *AppHandler) {
	e.GET("/sse", echo.WrapHandler(sse), sseConnectionLimiter(10))

	assets := e.Group("/assets", longCacheLifetime)
	assets.Static("/", "assets")

	icons := e.Group("/icons", longCacheLifetime)
	icons.Static("/", "config/icons")

	e.GET("/robots.txt", func(ctx *echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	e.GET("/", appHandler.handleIndex)

	e.GET("/*", func(c *echo.Context) error {
		return c.Redirect(http.StatusFound, "/")
	})
}
