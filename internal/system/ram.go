package system

import (
	"math"

	"github.com/dariubs/percent"
	"github.com/shirou/gopsutil/v3/mem"
)

func staticRam() Ram {
	var result = Ram{}
	r, err := mem.VirtualMemory()
	if err != nil {
		return result
	}
	result.Total = readableSize(r.Total)
	if r.SwapTotal > 0 {
		result.Swap = readableSize(r.SwapTotal) + " swap"
	} else {
		result.Swap = "No swap"
	}
	return result
}

func (c *Config) liveRam() {
	r, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	c.System.Live.Ram.Value = readableSize(r.Used)
	c.System.Live.Ram.Percentage = math.RoundToEven(percent.PercentOfFloat(float64(r.Used), float64(r.Total)))
}
