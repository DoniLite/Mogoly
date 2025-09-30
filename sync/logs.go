package sync

import "fmt"

var logger Logger

func SetLogger(lgr Logger) {
	logger = lgr
}

func getLogger() (Logger, error) {
	if logger == nil {
		return nil, fmt.Errorf("no logger assigned for this instance")
	}

	return logger, nil
}

func logf(level LogType, format string, args ...any) {
	currentLogger, err := getLogger()
	if currentLogger == nil || err != nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	select {
	case currentLogger <- Logs{Message: msg, LogType: level}:
	default:
	}
}