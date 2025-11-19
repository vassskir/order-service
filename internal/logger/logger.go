package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Service   string      `json:"service"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	serviceName string
	prettyPrint bool 
)


func Init(service string) {
	serviceName = service
	
	prettyPrint = os.Getenv("LOG_PRETTY") == "true"
	
	infoLogger = log.New(os.Stdout, "", 0)
	errorLogger = log.New(os.Stderr, "", 0)
	warnLogger = log.New(os.Stdout, "", 0)
}

func logInternal(level, message string, data interface{}, err error) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Service:   serviceName,
		Message:   message,
		Data:      data,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if prettyPrint {
		prettyLog(level, message, data, err)
		return
	}

	jsonData, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		log.Printf("[%s] %s: %v", level, message, err)
		return
	}

	switch level {
	case "ERROR":
		errorLogger.Println(string(jsonData))
	case "WARN":
		warnLogger.Println(string(jsonData))
	default:
		infoLogger.Println(string(jsonData))
	}
}

func prettyLog(level, message string, data interface{}, err error) {
	var color string
	switch level {
	case "ERROR":
		color = "🔴" 
	case "WARN":
		color = "🟡" 
	case "INFO":
		color = "🔵" 
	default:
		color = "⚪" 
	}

	log.Printf("%s [%s] %s", color, level, message)

	if data != nil {
		if dataMap, ok := data.(map[string]interface{}); ok {
			for key, value := range dataMap {
				log.Printf("   📍 %s: %v", key, value)
			}
		}
	}

	// Ошибка
	if err != nil {
		log.Printf("   ❌ Error: %v", err)
	}
}

func Info(message string, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	logInternal("INFO", message, logData, nil)
}

func Error(message string, err error, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	logInternal("ERROR", message, logData, err)
}

func Warn(message string, data ...interface{}) {
	var logData interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	logInternal("WARN", message, logData, nil)
}