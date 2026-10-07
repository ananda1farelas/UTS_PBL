package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// Logger is the shared application logger. It writes to both stdout and
// logs/app.log so output is visible in the terminal and kept for later.
var Logger *log.Logger

// InitLogger sets up the shared Logger. Call it once at startup, after Load().
func InitLogger() {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("config: failed to create log directory: %v", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, "app.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatalf("config: failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	Logger = log.New(multiWriter, "", log.Ldate|log.Ltime)
}
