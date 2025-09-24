package services

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"gitlab.unjx.de/flohoss/godash/internal/readable"
)

type SystemService struct {
	sse    *sse.Server
	mu     sync.Mutex
	buffer Buffer
}

type Buffer struct {
	CPUStr string `json:"cpu_str"`
	CPU    []CPU  `json:"cpu"`
	RAM    string `json:"ram"`
	Disk   string `json:"disk"`
}

type CPU struct {
	Time    int64 `json:"time"`
	Percent int   `json:"percent"`
}

func NewSystemService(sse *sse.Server) *SystemService {
	s := SystemService{
		sse: sse,
	}
	sse.CreateStream("system")
	go s.collect()
	return &s
}

func (s *SystemService) GetBuffer() *Buffer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.buffer
}

func (s *SystemService) collect() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

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
		cpu := int(math.Floor(cpuPercent[0]))
		s.buffer.CPU = append(s.buffer.CPU, CPU{
			Time:    time.Now().UnixMilli(),
			Percent: cpu,
		})
		if len(s.buffer.CPU) > 60 {
			s.buffer.CPU = s.buffer.CPU[1:]
		}

		s.buffer.CPUStr = fmt.Sprintf("%d %%", cpu)

		s.buffer.RAM = readable.ReadableSizePair(memStat.Used, memStat.Total)
		s.buffer.Disk = readable.ReadableSizePair(diskStat.Used, diskStat.Total)

		snapshot := s.buffer
		s.mu.Unlock()

		data, _ := json.Marshal(snapshot)
		s.sse.Publish("system", &sse.Event{Data: data})
	}
}
