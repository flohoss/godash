package controller

import (
	"github.com/r3labs/sse/v2"
	"gitlab.unjx.de/flohoss/godash/internal/bookmarks"
	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/internal/system"
	"gitlab.unjx.de/flohoss/godash/internal/weather"
)

type Controller struct {
	ENV  *env.Config
	SSE  *sse.Server
	Info Info
}

type Info struct {
	Weather   *weather.Weather
	Bookmarks *bookmarks.Config
	System    *system.System
}

func NewController(env *env.Config) *Controller {
	ctrl := Controller{
		ENV: env,
		SSE: sse.New(),
	}
	ctrl.SSE.AutoReplay = false
	ctrl.SSE.CreateStream("system")
	ctrl.SSE.CreateStream("weather")

	ctrl.Info = Info{
		Weather:   weather.NewWeatherService(ctrl.SSE),
		Bookmarks: bookmarks.NewBookmarkService(),
		System:    system.NewSystemService(ctrl.ENV.LiveSystem, ctrl.SSE),
	}
	return &ctrl
}
