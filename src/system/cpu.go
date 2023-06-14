package system

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"math"
	"runtime"
	"strconv"
)

func staticCpu() CPU {
	var p CPU
	p.Threads = strconv.Itoa(runtime.NumCPU()) + " threads"
	c, err := cpu.Info()
	if err == nil {
		p.Name = c[0].ModelName
	} else {
		p.Name = "none detected"
	}
	return p
}

func (c *Config) liveCpu() {
	p, err := cpu.Percent(0, false)
	if err != nil {
		return
	}
	c.System.Live.CPU = math.RoundToEven(p[0])
}
