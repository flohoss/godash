package system

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/r3labs/sse/v2"
)

func NewSystemService(enabled bool, sse *sse.Server) *System {
	var s Config
	if enabled {
		s = Config{sse: sse}
		s.Initialize()
	}
	return &s.System
}

func (c *Config) UpdateLiveInformation() {
	for {
		c.liveCpu()
		c.liveRam()
		c.liveDisk()
		c.uptime()
		json, _ := json.Marshal(c.System.Live)
		c.sse.Publish("system", &sse.Event{Data: json})
		time.Sleep(1 * time.Second)
	}
}

func (c *Config) Initialize() {
	c.System.Static.Host = staticHost()
	c.System.Static.CPU = staticCpu()
	c.System.Static.Ram = staticRam()
	c.System.Static.Disk = staticDisk()
	go c.UpdateLiveInformation()
	slog.Debug("system updated", "cpu", c.System.Static.CPU.Name, "arch", c.System.Static.Host.Architecture)
}
