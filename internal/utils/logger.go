package utils

import (
	"fmt"
	"log"
	"time"
)

// LogWithLabelAndTimestamp выводит сообщение с меткой и временем.
func LogWithLabelAndTimestamp(label, text string) {
	log.Println(fmt.Sprintf("[%s] %s | %s", label, time.Now().Format("15:04:05"), text))
}

// TruncateForLog обрезает строку до max байт для вывода в лог; хвост заменяется на «…».
func TruncateForLog(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
