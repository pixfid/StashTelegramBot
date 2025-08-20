package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// Logger - обертка для красивого логирования
type Logger struct {
	prefix string
}

func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

func (l *Logger) Info(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	color.Cyan("[%s] [%s INFO] %s", timestamp, l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Success(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	color.Green("[%s] [%s SUCCESS] ✓ %s", timestamp, l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	color.Red("[%s] [%s ERROR] ✗ %s", timestamp, l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Warning(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	color.Yellow("[%s] [%s WARN] ⚠ %s", timestamp, l.prefix, fmt.Sprintf(format, args...))
}
