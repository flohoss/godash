package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
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

func NewAppHandler(env *env.Config, store *sessions.CookieStore, s SystemService, w WeatherService, b BookmarkService) *AppHandler {
	return &AppHandler{
		env:             env,
		store:           store,
		systemService:   s,
		weatherService:  w,
		bookmarkService: b,
	}
}

type AppHandler struct {
	env             *env.Config
	store           *sessions.CookieStore
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

	session, _ := bh.store.Get(r, StoreSessionKey)
	user := &services.User{
		Name:     session.Values[string(NameKey)].(string),
		Email:    session.Values[string(EmailKey)].(string),
		Gravatar: session.Values[string(GravatarKey)].(string),
	}

	home.HomeIndex(titlePage, bh.env.Version, home.Home(titlePage, user, bookmarks, staticSystem, liveSystem, weather)).Render(r.Context(), w)
}
