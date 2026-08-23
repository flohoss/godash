package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/r3labs/sse/v2"
)

func TestSystemServicePublishSnapshot(t *testing.T) {
	s := sse.New()
	s.CreateStream("system")
	srv := httptest.NewServer(http.HandlerFunc(s.ServeHTTP))
	defer srv.Close()
	defer s.Close()

	svc := &SystemService{
		sse: s,
		buffer: Buffer{
			CPU:  Detail{Value: "Intel", Percentage: 10},
			RAM:  Detail{Value: "8 GB", Percentage: 50},
			Disk: Detail{Value: "100 GB", Percentage: 40},
		},
	}

	client := sse.NewClient(srv.URL)
	events := make(chan string, 4)
	go func() {
		_ = client.Subscribe("system", func(msg *sse.Event) {
			if len(msg.Data) > 0 {
				events <- string(msg.Data)
			}
		})
	}()
	time.Sleep(300 * time.Millisecond)

	svc.publishSnapshot()

	// Expect 5 publishes (cpu-percentage, ram-value, ram-percentage, disk-value, disk-percentage).
	for i := 0; i < 5; i++ {
		select {
		case <-events:
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for system event %d", i)
		}
	}
}

func TestSystemServiceGetters(t *testing.T) {
	svc := &SystemService{
		static: Static{CPU: "8 threads", RAM: "no swap", Disk: "1 partitions"},
		buffer: Buffer{CPU: Detail{Value: "Intel", Percentage: 10}},
	}
	if got := svc.GetStatic(); got.CPU != "8 threads" {
		t.Errorf("GetStatic().CPU = %q", got.CPU)
	}
	if got := svc.GetBuffer(); got.CPU.Percentage != 10 {
		t.Errorf("GetBuffer().CPU.Percentage = %d", got.CPU.Percentage)
	}
}
