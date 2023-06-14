package system

import (
	"encoding/json"
	"time"

	"github.com/r3labs/sse/v2"
	"go.uber.org/zap"
)

func NewSystemService(enabled bool, logging *zap.SugaredLogger, sse *sse.Server) *System {
	var s Config
	if enabled {
		s = Config{log: logging, sse: sse}
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
	c.log.Debugw("system updated", "cpu", c.System.Static.CPU.Name, "arch", c.System.Static.Host.Architecture)
}
