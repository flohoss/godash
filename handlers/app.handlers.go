package handlers

import (
	"net/http"

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

func NewAppHandler(env *env.Config, s SystemService, w WeatherService, b BookmarkService) *AppHandler {
	return &AppHandler{
		env:             env,
		systemService:   s,
		weatherService:  w,
		bookmarkService: b,
	}
}

type AppHandler struct {
	env             *env.Config
	systemService   SystemService
	weatherService  WeatherService
	bookmarkService BookmarkService
}

func (bh *AppHandler) appHandler(w http.ResponseWriter, r *http.Request) {
	bookmarks := bh.bookmarkService.GetAllBookmarks()
	staticSystem := bh.systemService.GetStaticInformation()
	liveSystem := bh.systemService.GetLiveInformation()
	weather := bh.weatherService.GetCurrentWeather()

	titlePage := bh.env.Title

	user := services.User{
		Name:  w.Header().Get("X-User-Name"),
		Email: w.Header().Get("X-User-Email"),
	}
	gravatar := services.NewGravatarFromEmail(user.Email)
	user.Gravatar = gravatar.GetURL()

	home.HomeIndex(titlePage, bh.env.Version, home.Home(titlePage, user, bookmarks, staticSystem, liveSystem, weather)).Render(r.Context(), w)
}
