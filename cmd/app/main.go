package main

import (
	"log/slog"

	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

// TODO: application logic
func main() {
	logger.Init(slog.LevelDebug)
	logger.Log.Info("Starting application...")
}