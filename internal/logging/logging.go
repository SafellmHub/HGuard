package logging

import (
	"log"
)

func Info(msg string, args ...interface{}) {
	log.Printf("INFO: "+msg, args...)
}

func Warn(msg string, args ...interface{}) {
	log.Printf("WARN: "+msg, args...)
}

func Error(msg string, args ...interface{}) {
	log.Printf("ERROR: "+msg, args...)
}

// Placeholder for logging setup
