package router

import (
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.unjx.de/flohoss/godash/internal/controller"
)

func InitRouter() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(echo.WrapMiddleware(chiMiddleware.Heartbeat("/health")))
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	e.Renderer = initTemplates()

	return e
}

func SetupRoutes(e *echo.Echo, ctrl *controller.Controller) {
	static := e.Group("/static", longCacheLifetime)
	static.Static("/", "web/static")

	storage := e.Group("/storage", longCacheLifetime)
	storage.Static("/icons", "storage/icons")

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Title":   ctrl.ENV.Title,
			"Weather": ctrl.Info.Weather.CurrentWeather,
			"Parsed":  ctrl.Info.Bookmarks.Parsed,
			"System":  ctrl.Info.System,
		})
	})

	e.GET("/sse", echo.WrapHandler(http.HandlerFunc(ctrl.SSE.ServeHTTP)))

	e.GET("/robots.txt", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})
	e.RouteNotFound("*", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/")
	})
}
