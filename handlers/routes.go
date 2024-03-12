package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
)

func SetupRoutes(e *echo.Echo, sse *sse.Server, bh *AppHandler) {
	e.GET("/", bh.appHandler)
	e.GET("/sse", echo.WrapHandler(http.HandlerFunc(sse.ServeHTTP)))
}

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
