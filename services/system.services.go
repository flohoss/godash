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
	mu     sync.RWMutex
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
	RegisterSnapshot("system", s.publishSnapshot)
	go s.collect()
	return &s
}

func (s *SystemService) publishString(id string, v string) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.sse.Publish("system", &sse.Event{Event: []byte(id), Data: append([]byte(nil), data...)})
}

func (s *SystemService) publishInt(id string, n int) {
	data, err := json.Marshal(n)
	if err != nil {
		return
	}
	s.sse.Publish("system", &sse.Event{Event: []byte(id), Data: append([]byte(nil), data...)})
}
func (s *SystemService) publishSnapshot() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.publishInt("cpu-percentage", s.buffer.CPU.Percentage)
	s.publishString("ram-value", s.buffer.RAM.Value)
	s.publishInt("ram-percentage", s.buffer.RAM.Percentage)
	s.publishString("disk-value", s.buffer.Disk.Value)
	s.publishInt("disk-percentage", s.buffer.Disk.Percentage)
}
func (s *SystemService) GetBuffer() Buffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buffer
}

func (s *SystemService) GetStatic() Static {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.static
}

func (s *SystemService) computeStatic() Static {
	var st Static
	st.CPU = strconv.Itoa(runtime.NumCPU()) + " threads"
	p, err := disk.Partitions(false)
	if err == nil {
		st.Disk = strconv.Itoa(len(p)) + " partitions"
	}

	r, err := mem.VirtualMemory()
	if err == nil && r.SwapTotal > 0 {
		st.RAM = readable.ReadableSize(r.SwapTotal) + " swap"
	} else {
		st.RAM = "no swap"
	}
	return st
}

func (s *SystemService) collect() {
	cpuTicker := time.NewTicker(time.Second)
	defer cpuTicker.Stop()
	diskTicker := time.NewTicker(10 * time.Second)
	defer diskTicker.Stop()

	c, err := cpu.Info()
	cpuModel := ""
	if err == nil && len(c) > 0 {
		if c[0].ModelName != "" {
			cpuModel = c[0].ModelName
		} else {
			cpuModel = c[0].VendorID
		}
	}
	static := s.computeStatic()

	s.mu.Lock()
	s.buffer.CPU.Value = cpuModel
	s.static = static
	s.mu.Unlock()

	cpu.Percent(time.Second, false)

	var prevCPUPct, prevRAMPct, prevDiskPct int
	var prevRAMVal, prevDiskVal string

	pollDisk := func() {
		diskStat, err := disk.Usage("/")
		if err != nil {
			return
		}
		newDiskPct := int(math.Floor(diskStat.UsedPercent))
		newDiskVal := readable.ReadableSizePair(diskStat.Used, diskStat.Total)

		var publishes []func()
		s.mu.Lock()
		if newDiskVal != prevDiskVal {
			prevDiskVal = newDiskVal
			s.buffer.Disk.Value = newDiskVal
			val := newDiskVal
			publishes = append(publishes, func() { s.publishString("disk-value", val) })
		}
		if newDiskPct != prevDiskPct {
			prevDiskPct = newDiskPct
			s.buffer.Disk.Percentage = newDiskPct
			pct := newDiskPct
			publishes = append(publishes, func() { s.publishInt("disk-percentage", pct) })
		}
		s.mu.Unlock()
		for _, fn := range publishes {
			fn()
		}
	}

	pollDisk()

	for {
		select {
		case <-cpuTicker.C:
			cpuPercent, err := cpu.Percent(0, false)
			if err != nil || len(cpuPercent) == 0 {
				continue
			}

			memStat, err := mem.VirtualMemory()
			if err != nil {
				continue
			}

			newCPUPct := int(math.Floor(cpuPercent[0]))
			newRAMPct := int(math.Floor(memStat.UsedPercent))
			newRAMVal := readable.ReadableSizePair(memStat.Used, memStat.Total)

			var publishes []func()

			s.mu.Lock()

			if newCPUPct != prevCPUPct {
				prevCPUPct = newCPUPct
				s.buffer.CPU.Percentage = newCPUPct
				pct := newCPUPct
				publishes = append(publishes, func() { s.publishInt("cpu-percentage", pct) })
			}

			if newRAMVal != prevRAMVal {
				prevRAMVal = newRAMVal
				s.buffer.RAM.Value = newRAMVal
				val := newRAMVal
				publishes = append(publishes, func() { s.publishString("ram-value", val) })
			}

			if newRAMPct != prevRAMPct {
				prevRAMPct = newRAMPct
				s.buffer.RAM.Percentage = newRAMPct
				pct := newRAMPct
				publishes = append(publishes, func() { s.publishInt("ram-percentage", pct) })
			}

			s.mu.Unlock()

			for _, fn := range publishes {
				fn()
			}

		case <-diskTicker.C:
			pollDisk()
		}
	}
}
