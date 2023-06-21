package system

import (
	"github.com/r3labs/sse/v2"
	"go.uber.org/zap"
)

type Config struct {
	sse    *sse.Server
	log    *zap.SugaredLogger
	System System
}

type System struct {
	Live   LiveInformation   `json:"live"`
	Static StaticInformation `json:"static"`
}

type LiveStorageInformation struct {
	Value      string  `json:"value"`
	Percentage float64 `json:"percentage"`
}

type LiveInformation struct {
	CPU    float64                `json:"cpu"`
	Ram    LiveStorageInformation `json:"ram"`
	Disk   LiveStorageInformation `json:"disk"`
	Uptime Uptime                 `json:"uptime"`
}

type Uptime struct {
	Days       uint64  `json:"days"`
	Hours      uint16  `json:"hours"`
	Minutes    uint16  `json:"minutes"`
	Seconds    uint16  `json:"seconds"`
	Percentage float32 `json:"percentage"`
}

type CPU struct {
	Name    string `json:"name"`
	Threads string `json:"threads"`
}

type Host struct {
	Architecture string `json:"architecture"`
}

type Ram struct {
	Total string `json:"total"`
	Swap  string `json:"swap"`
}

type Disk struct {
	Total      string `json:"total"`
	Partitions string `json:"partitions"`
}

type StaticInformation struct {
	CPU  CPU  `json:"cpu"`
	Ram  Ram  `json:"ram"`
	Disk Disk `json:"disk"`
	Host Host `json:"host"`
}
