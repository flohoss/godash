package services

import (
	"encoding/json"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/flohoss/godash/internal/readable"
	"github.com/r3labs/sse/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemService struct {
	sse    *sse.Server
	mu     sync.Mutex
	static Static
	buffer Buffer
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

func NewSystemService(sse *sse.Server) *SystemService {
	s := SystemService{sse: sse}
	sse.CreateStream("system")
	go s.collect()
	RegisterSnapshot("system", s.publishSnapshot)
	return &s
}

func (s *SystemService) publishJSON(id string, d Detail) {
	data, err := json.Marshal(d)
	if err != nil {
		return
	}
	s.sse.Publish("system", &sse.Event{Event: []byte(id), Data: append([]byte(nil), data...)})
}

func (s *SystemService) publishSnapshot() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.publishJSON("cpu", s.buffer.CPU)
	s.publishJSON("ram", s.buffer.RAM)
	s.publishJSON("disk", s.buffer.Disk)
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

		newCPU := int(math.Floor(cpuPercent[0]))
		newRAM := int(math.Floor(memStat.UsedPercent))
		newDisk := int(math.Floor(diskStat.UsedPercent))

		s.mu.Lock()

		if newCPU != prevCPU {
			prevCPU = newCPU
			s.buffer.CPU.Percentage = newCPU
			s.publishJSON("cpu", s.buffer.CPU)
		}

		if newRAM != prevRAM {
			prevRAM = newRAM
			s.buffer.RAM = Detail{
				Value:      readable.ReadableSizePair(memStat.Used, memStat.Total),
				Percentage: newRAM,
			}
			s.publishJSON("ram", s.buffer.RAM)
		}

		if newDisk != prevDisk {
			prevDisk = newDisk
			s.buffer.Disk = Detail{
				Value:      readable.ReadableSizePair(diskStat.Used, diskStat.Total),
				Percentage: newDisk,
			}
			s.publishJSON("disk", s.buffer.Disk)
		}

		s.mu.Unlock()
	}
}
