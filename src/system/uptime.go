package system

import (
	"github.com/shirou/gopsutil/v3/host"
)

func (c *Config) uptime() {
	i, err := host.Info()
	if err != nil {
		return
	}
	c.System.Live.Uptime.Days = i.Uptime / 84600
	c.System.Live.Uptime.Hours = uint16((i.Uptime % 86400) / 3600)
	c.System.Live.Uptime.Minutes = uint16(((i.Uptime % 86400) % 3600) / 60)
	c.System.Live.Uptime.Seconds = uint16(((i.Uptime % 86400) % 3600) % 60)
	c.System.Live.Uptime.Percentage = float32((c.System.Live.Uptime.Minutes*100)+c.System.Live.Uptime.Seconds) / 60
}
