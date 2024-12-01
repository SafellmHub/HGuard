package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

var logger = log.New(os.Stdout, "", log.LstdFlags)

func logStructured(level, msg, requestID string, fields map[string]interface{}) {
	entry := LogEntry{
		Level:     level,
		Message:   msg,
		RequestID: requestID,
		Fields:    fields,
	}
	b, _ := json.Marshal(entry)
	logger.Println(string(b))
}

func Info(msg string, args ...interface{}) {
	logStructured("info", format(msg, args...), "", nil)
}

func Warn(msg string, args ...interface{}) {
	logStructured("warn", format(msg, args...), "", nil)
}

func Error(msg string, args ...interface{}) {
	logStructured("error", format(msg, args...), "", nil)
}

func InfoWithID(requestID, msg string, fields map[string]interface{}) {
	logStructured("info", msg, requestID, fields)
}

func WarnWithID(requestID, msg string, fields map[string]interface{}) {
	logStructured("warn", msg, requestID, fields)
}

func ErrorWithID(requestID, msg string, fields map[string]interface{}) {
	logStructured("error", msg, requestID, fields)
}

func format(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}
