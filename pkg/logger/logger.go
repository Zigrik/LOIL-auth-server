package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[AUTH] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.Println("[INFO]", v)
}

func (l *Logger) Error(v ...interface{}) {
	l.Println("[ERROR]", v)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Println("[FATAL]", v)
	os.Exit(1)
}