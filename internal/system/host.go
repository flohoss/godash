package system

import (
	"runtime"
)

func staticHost() Host {
	var h Host
	h.Architecture = runtime.GOARCH
	return h
}
