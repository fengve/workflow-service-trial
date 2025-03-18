package core

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
)

var (
	Logger = log.With(log.NewJSONLogger(os.Stdout), "app", "workflow-service")
)

func Warnf(format string, args ...interface{}) {
	Logger.Log("Warn", fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}) {
	Logger.Log("Debug", fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	Logger.Log("Error", fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	Logger.Log("Info", fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	Logger.Log("Fatal", fmt.Sprintf(format, args...))
}

func Panicf(format string, args ...interface{}) {
	Logger.Log("Panic", fmt.Sprintf(format, args...))
}
