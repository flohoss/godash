package handlers

import (
	"github.com/labstack/echo/v4"
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

func NewAppHandler(env *env.Config, s SystemService, b BookmarkService) *AppHandler {

	return &AppHandler{
		env:             env,
		SystemService:   s,
		BookmarkService: b,
	}
}

type AppHandler struct {
	env             *env.Config
	SystemService   SystemService
	BookmarkService BookmarkService
}

func (bh *AppHandler) appHandler(c echo.Context) error {
	bookmarks := bh.BookmarkService.GetAllBookmarks()
	staticSystem := bh.SystemService.GetStaticInformation()
	liveSystem := bh.SystemService.GetLiveInformation()

	titlePage := bh.env.Title

	return renderView(c, home.HomeIndex(titlePage, bh.env.Version, bookmarks, staticSystem, liveSystem, home.Home(titlePage, bookmarks, staticSystem, liveSystem)))
}
