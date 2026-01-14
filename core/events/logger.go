package events

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/DoniLite/Mogoly/core/config"
	goevents "github.com/DoniLite/go-events"
)

var (
	logger *Logger
	logChannel LogChannel
)

func init() {
	logChannel = LogChannel{ch: make(chan any, 100)}
	logger = newLogger()
}

func newLogger() *Logger {
	logFile, err := createLogFile()
	if err != nil {
		return &Logger{writer: io.MultiWriter(os.Stdout, &logChannel)}
	}
	return &Logger{writer: io.MultiWriter(os.Stdout, logFile, &logChannel)}
}

func createLogFile() (*os.File, error) {
	logFilePath := config.GetEnv(config.LOG_FILE, "mogoly.log")
	logPath, err := config.CreateConfigFile(logFilePath)
	if err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func Logf(level LogType, format string, args ...any) {
	currentLogger := GetLogger()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if currentLogger == nil {
		return
	}
	msg := fmt.Sprintf("=====================\n[%s] %s \n=====================\n\n", timestamp, fmt.Sprintf(format, args...))
	currentLogger.writer.Write([]byte(msg))
	switch level {
	case LOG_ERROR:
		GetEventBus().Emit(ErrorDroppedEvent, &goevents.EventData{
			Message: "Error Droped",
			Payload: fmt.Sprintf(format, args...),
		})
	}
}

func GetLogger() *Logger {
	if logger == nil {
		logger = newLogger()
		return logger
	}

	return logger
}

func SetLogger(lgr *Logger) {
	logger = lgr
}

func GetLogChannel() chan any {
	return logChannel.ch
}

func StreamLogs() <-chan any {
	return logChannel.ch
}

func GetLogs() []any {
	logs := make([]any, 0)
	for log := range logChannel.ch {
		logs = append(logs, log)
	}
	return logs
}
