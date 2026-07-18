package services

import "sync"

var (
	snapshotMu       sync.RWMutex
	snapshotByStream = map[string]func(){}
)

func RegisterSnapshot(streamID string, fn func()) {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	snapshotByStream[streamID] = fn
}

func PublishSnapshot(streamID string) {
	snapshotMu.RLock()
	fn := snapshotByStream[streamID]
	snapshotMu.RUnlock()
	if fn != nil {
		fn()
	}
}
