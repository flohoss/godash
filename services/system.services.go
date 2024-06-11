package services

import (
	"encoding/json"
	"math"
	"runtime"
	"strconv"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"gitlab.unjx.de/flohoss/godash/internal/readable"
)

func calculatePercentage(used, total uint64) float64 {
	if total == 0 {
		return 0.0
	}
	return (float64(used) / float64(total)) * 100
}

func NewSystemService(sse *sse.Server) *SystemService {
	s := SystemService{
		sse: sse,
		Static: StaticInformation{
			CPU:  staticCpu(),
			Ram:  staticRam(),
			Disk: staticDisk(),
			Host: staticHost(),
		},
	}
	sse.CreateStream("system")
	go s.UpdateLiveInformation()
	return &s
}

func (s *SystemService) GetLiveInformation() *LiveInformation {
	return &s.Live
}

func (s *SystemService) GetStaticInformation() *StaticInformation {
	return &s.Static
}

func (s *SystemService) UpdateLiveInformation() {
	for {
		s.liveCpu()
		s.liveRam()
		s.liveDisk()
		s.uptime()
		json, _ := json.Marshal(s.Live)
		s.sse.Publish("system", &sse.Event{Data: json})
		time.Sleep(1 * time.Second)
	}
}

func staticHost() Host {
	var h Host
	h.Architecture = runtime.GOARCH
	return h
}

func staticCpu() CPU {
	var p CPU
	p.Threads = strconv.Itoa(runtime.NumCPU()) + " threads"
	c, err := cpu.Info()
	if err == nil {
		if c[0].ModelName != "" {
			p.Name = c[0].ModelName
		} else {
			p.Name = c[0].VendorID
		}
	} else {
		p.Name = "none detected"
	}
	return p
}

func (s *SystemService) liveCpu() {
	p, err := cpu.Percent(0, false)
	if err != nil {
		return
	}
	s.Live.CPU = math.RoundToEven(p[0])
}

func staticRam() Ram {
	var result = Ram{}
	r, err := mem.VirtualMemory()
	if err != nil {
		return result
	}
	result.Total = readable.ReadableSize(r.Total)
	if r.SwapTotal > 0 {
		result.Swap = readable.ReadableSize(r.SwapTotal) + " swap"
	} else {
		result.Swap = "no swap"
	}
	return result
}

func (s *SystemService) liveRam() {
	r, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	s.Live.Ram.Value = readable.ReadableSize(r.Used)
	s.Live.Ram.Percentage = math.RoundToEven(calculatePercentage(r.Used, r.Total))
}

func staticDisk() Disk {
	var result = Disk{}
	d, err := disk.Usage("/")
	if err != nil {
		return result
	}
	p, err := disk.Partitions(false)
	if err != nil {
		return result
	}
	result.Total = readable.ReadableSize(d.Total)
	result.Partitions = strconv.Itoa(len(p)) + " partitions"
	return result
}

func (s *SystemService) liveDisk() {
	d, err := disk.Usage("/")
	if err != nil {
		return
	}
	s.Live.Disk.Value = readable.ReadableSize(d.Used)
	s.Live.Disk.Percentage = math.RoundToEven(calculatePercentage(d.Used, d.Total))
}

func (s *SystemService) uptime() {
	i, err := host.Info()
	if err != nil {
		return
	}
	s.Live.Uptime.Days = i.Uptime / 84600
	s.Live.Uptime.Hours = uint16((i.Uptime % 86400) / 3600)
	s.Live.Uptime.Minutes = uint16(((i.Uptime % 86400) % 3600) / 60)
	s.Live.Uptime.Seconds = uint16(((i.Uptime % 86400) % 3600) % 60)
	s.Live.Uptime.Percentage = float32((s.Live.Uptime.Minutes*100)+s.Live.Uptime.Seconds) / 60
}
