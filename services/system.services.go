package services

import (
	"bytes"
	"context"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/flohoss/godash/internal/readable"
	"github.com/r3labs/sse/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemService struct {
	sse         *sse.Server
	mu          sync.Mutex
	static      Static
	buffer      Buffer
	renderBadge func(string, string, string, Detail) templ.Component
}

type Static struct {
	CPU  string `json:"cpu"`
	RAM  string `json:"ram"`
	Disk string `json:"disk"`
}

type Buffer struct {
	CPU  Detail `json:"cpu"`
	RAM  Detail `json:"ram"`
	Disk Detail `json:"disk"`
}

type Detail struct {
	Value      string `json:"value"`
	Percentage int    `json:"percentage"`
}

func NewSystemService(sse *sse.Server, renderBadge func(string, string, string, Detail) templ.Component) *SystemService {
	s := SystemService{sse: sse, renderBadge: renderBadge}
	sse.CreateStream("system")
	go s.collect()
	return &s
}

func (s *SystemService) GetBuffer() *Buffer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.buffer
}

func (s *SystemService) GetStatic() *Static {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.static
}

func (s *SystemService) initStatic() {
	s.static.CPU = strconv.Itoa(runtime.NumCPU()) + " threads"
	p, err := disk.Partitions(false)
	if err == nil {
		s.static.Disk = strconv.Itoa(len(p)) + " partitions"
	}

	r, err := mem.VirtualMemory()
	if err == nil && r.SwapTotal > 0 {
		s.static.RAM = readable.ReadableSize(r.SwapTotal) + " swap"
	} else {
		s.static.RAM = "no swap"
	}
}

func (s *SystemService) collect() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	s.mu.Lock()
	c, err := cpu.Info()
	if err == nil {
		if c[0].ModelName != "" {
			s.buffer.CPU.Value = c[0].ModelName
		} else {
			s.buffer.CPU.Value = c[0].VendorID
		}
	}
	s.initStatic()
	s.mu.Unlock()

	// Reusable buffers to reduce allocations
	var cpuBuf, ramBuf, diskBuf bytes.Buffer

	// Track previous values to avoid unnecessary updates
	var prevCPU, prevRAM, prevDisk int

	for range ticker.C {
		cpuPercent, err := cpu.Percent(0, false)
		if err != nil || len(cpuPercent) == 0 {
			continue
		}

		memStat, err := mem.VirtualMemory()
		if err != nil {
			continue
		}

		diskStat, err := disk.Usage("/")
		if err != nil {
			continue
		}

		// Calculate new values
		newCPU := int(math.Floor(cpuPercent[0]))
		newRAM := int(math.Floor(memStat.UsedPercent))
		newDisk := int(math.Floor(diskStat.UsedPercent))

		s.mu.Lock()

		// Update and publish CPU (changes frequently)
		if newCPU != prevCPU {
			prevCPU = newCPU
			s.buffer.CPU.Percentage = newCPU

			cpuBuf.Reset()
			if err := s.renderBadge("cpu", "icon-[bi--cpu]", s.static.CPU, s.buffer.CPU).Render(context.Background(), &cpuBuf); err == nil {
				s.sse.Publish("system", &sse.Event{Event: []byte("cpu"), Data: cpuBuf.Bytes()})
			}
		}

		// Update and publish RAM (changes less frequently)
		if newRAM != prevRAM {
			prevRAM = newRAM
			s.buffer.RAM = Detail{
				Value:      readable.ReadableSizePair(memStat.Used, memStat.Total),
				Percentage: newRAM,
			}

			ramBuf.Reset()
			if err := s.renderBadge("ram", "icon-[bi--memory]", s.static.RAM, s.buffer.RAM).Render(context.Background(), &ramBuf); err == nil {
				s.sse.Publish("system", &sse.Event{Event: []byte("ram"), Data: ramBuf.Bytes()})
			}
		}

		// Update and publish Disk (changes rarely)
		if newDisk != prevDisk {
			prevDisk = newDisk
			s.buffer.Disk = Detail{
				Value:      readable.ReadableSizePair(diskStat.Used, diskStat.Total),
				Percentage: newDisk,
			}

			diskBuf.Reset()
			if err := s.renderBadge("disk", "icon-[bi--hdd]", s.static.Disk, s.buffer.Disk).Render(context.Background(), &diskBuf); err == nil {
				s.sse.Publish("system", &sse.Event{Event: []byte("disk"), Data: diskBuf.Bytes()})
			}
		}

		s.mu.Unlock()
	}
}
