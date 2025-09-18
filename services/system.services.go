package services

import (
	"encoding/json"
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
	buffer []UsagePoint
}

type UsagePoint struct {
	CPU  float64 `json:"cpu"`
	Ram  string  `json:"ram"`
	Disk string  `json:"disk"`
}

func NewSystemService(sse *sse.Server) *SystemService {
	s := SystemService{
		sse: sse,
	}
	sse.CreateStream("system")
	go s.collect()
	return &s
}

func (s *SystemService) GetBuffer() []UsagePoint {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := make([]UsagePoint, len(s.buffer))
	copy(snapshot, s.buffer)
	return snapshot
}

func (s *SystemService) collect() {
	ticker := time.NewTicker(time.Second)
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

		point := UsagePoint{
			CPU:  cpuPercent[0],
			Ram:  readable.ReadableSizePair(memStat.Used, memStat.Total),
			Disk: readable.ReadableSizePair(diskStat.Used, diskStat.Total),
		}

		s.mu.Lock()
		s.buffer = append(s.buffer, point)
		if len(s.buffer) > 60 {
			s.buffer = s.buffer[1:]
		}
		snapshot := make([]UsagePoint, len(s.buffer))
		copy(snapshot, s.buffer)
		s.mu.Unlock()

		data, _ := json.Marshal(snapshot[len(snapshot)-1])
		s.sse.Publish("system", &sse.Event{Data: data})
	}
}
