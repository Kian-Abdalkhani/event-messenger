package logger

import (
	"log/slog"
	"os"
)

func InitLogger() {

	// Set log level based on environment, default to Info
	debug := os.Getenv("DEBUG")
	var logLevel slog.Level
	if debug == "true" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	// Configure log handler (can be changed to JSON)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Set as default logger
	slog.SetDefault(logger)

}
