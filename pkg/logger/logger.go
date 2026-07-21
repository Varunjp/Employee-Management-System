package logger

import (
	"log"
	"os"
)

var std = log.New(os.Stdout, "", log.LstdFlags)

func Info(msg string, args ...any) {
	std.Printf("[INFO] "+msg, args...)
}

func Error(msg string, args ...any) {
	std.Printf("[ERROR] "+msg, args...)
}

func Fatal(msg string, args ...any) {
	std.Fatalf("[FATAL] "+msg, args...)
}
