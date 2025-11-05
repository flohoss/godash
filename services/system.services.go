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
	sse          *sse.Server
	mu           sync.Mutex
	static       Static
	buffer       Buffer
	renderReport func(*Buffer, *Static) templ.Component
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

func NewSystemService(sse *sse.Server, renderReport func(*Buffer, *Static) templ.Component) *SystemService {
	s := SystemService{sse: sse, renderReport: renderReport}
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
	if err != nil && r.SwapTotal > 0 {
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

		s.mu.Lock()
		s.buffer.CPU.Percentage = int(math.Floor(cpuPercent[0]))
		s.buffer.RAM = Detail{
			Value:      readable.ReadableSizePair(memStat.Used, memStat.Total),
			Percentage: int(math.Floor(memStat.UsedPercent)),
		}
		s.buffer.Disk = Detail{
			Value:      readable.ReadableSizePair(diskStat.Used, diskStat.Total),
			Percentage: int(math.Floor(diskStat.UsedPercent)),
		}
		s.mu.Unlock()

		var buf bytes.Buffer
		err = s.renderReport(s.GetBuffer(), s.GetStatic()).Render(context.Background(), &buf)
		if err == nil {
			s.sse.Publish("system", &sse.Event{Data: buf.Bytes()})
		}
	}
}
