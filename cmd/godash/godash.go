package main

import (
	"log"

	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/internal/logging"
	"go.uber.org/zap"
)

func main() {
	env, err := env.Parse()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logging.CreateLogger(env.LogLevel))
}
