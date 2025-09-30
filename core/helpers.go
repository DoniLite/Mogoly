package core

import "fmt"

func logf(level LogType, format string, args ...any) {
	currentLogger := GetLogger()
	if currentLogger == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	select {
	case currentLogger <- Logs{Message: msg, LogType: level}:
	default:
	}
}
