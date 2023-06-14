package main

import (
	"go.uber.org/zap"
)

func (g *goDash) setupLogger() {
	zapLevel, err := zap.ParseAtomicLevel(g.config.LogLevel)
	if err != nil {
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	g.logger = zap.Must(zap.Config{
		Level:            zapLevel,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}.Build()).Sugar()
}

func (g *goDash) setupEchoLogging() {
	g.router.HideBanner = true
	g.router.HidePort = true
}
