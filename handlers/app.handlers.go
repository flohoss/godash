package handlers

import (
	"log/slog"
	"net/http"

	"github.com/logto-io/go/client"
	"github.com/logto-io/go/core"

	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/services"
	"gitlab.unjx.de/flohoss/godash/views/home"
)

type BookmarkService interface {
	GetAllBookmarks() *services.Bookmarks
}

type SystemService interface {
	GetLiveInformation() *services.LiveInformation
	GetStaticInformation() *services.StaticInformation
}

type WeatherService interface {
	GetCurrentWeather() *services.OpenWeather
}

func NewAppHandler(env *env.Config, authHandler *AuthHandler, s SystemService, w WeatherService, b BookmarkService) *AppHandler {
	return &AppHandler{
		env:             env,
		authHandler:     authHandler,
		systemService:   s,
		weatherService:  w,
		bookmarkService: b,
	}
}

type AppHandler struct {
	env             *env.Config
	authHandler     *AuthHandler
	systemService   SystemService
	weatherService  WeatherService
	bookmarkService BookmarkService
}

func (bh *AppHandler) appHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	bookmarks := bh.bookmarkService.GetAllBookmarks()
	staticSystem := bh.systemService.GetStaticInformation()
	liveSystem := bh.systemService.GetLiveInformation()
	weather := bh.weatherService.GetCurrentWeather()

	var claims *core.IdTokenClaims
	if bh.authHandler.sessionManager != nil {
		logtoClient := client.NewLogtoClient(
			bh.authHandler.logtoConfig,
			&SessionStorage{
				sessionManager: bh.authHandler.sessionManager,
				write:          w,
				request:        r,
			},
		)
		c, err := logtoClient.GetIdTokenClaims()
		if err != nil {
			slog.Warn("cannot get id token claims", "err", err)
		}
		claims = &c
	}

	titlePage := bh.env.Title

	home.HomeIndex(titlePage, bh.env.Version, home.Home(titlePage, claims, bookmarks, staticSystem, liveSystem, weather)).Render(r.Context(), w)
}
