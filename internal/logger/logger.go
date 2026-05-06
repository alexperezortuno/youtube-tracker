package logger

import (
	"log"
	"os"
	"strings"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var currentLevel = INFO

func Init(level string) {
	switch strings.ToLower(level) {
	case "debug":
		currentLevel = DEBUG
	case "info":
		currentLevel = INFO
	case "warn":
		currentLevel = WARN
	case "error":
		currentLevel = ERROR
	default:
		currentLevel = INFO
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
}

func Debug(msg string, args ...any) {
	if currentLevel <= DEBUG {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func Info(msg string, args ...any) {
	if currentLevel <= INFO {
		log.Printf("[INFO] "+msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if currentLevel <= WARN {
		log.Printf("[WARN] "+msg, args...)
	}
}

func Error(msg string, args ...any) {
	if currentLevel <= ERROR {
		log.Printf("[ERROR] "+msg, args...)
	}
}
