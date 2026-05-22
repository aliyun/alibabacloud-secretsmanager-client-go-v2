package logger

import (
	"fmt"
	"os"
)

type Wrapper interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

type DefaultLogger struct {
}

func NewDefaultLogger() Wrapper {
	return &DefaultLogger{}
}

func (d DefaultLogger) Debug(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "[Debug] "+format+"\n", v...)
}

func (d DefaultLogger) Info(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "[Info] "+format+"\n", v...)
}

func (d DefaultLogger) Warn(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "[Warn] "+format+"\n", v...)
}

func (d DefaultLogger) Error(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "[Error] "+format+"\n", v...)
}
