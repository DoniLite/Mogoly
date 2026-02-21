package events

import "io"

type Logs struct {
	LogType LogType
	Message string
}

type LogType = int

const (
	LOG_INFO LogType = iota
	LOG_DEBUG
	LOG_ERROR
)

type Logger struct {
	writer io.Writer
}

type LoggerInterface interface {
	Log(logType LogType, message string)
}

type LogChannel struct {
	ch chan any
}
