package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

type Logger struct {
	Enabled bool
}

func (l *Logger) Printf(format string, args ...interface{}) {
	if l.Enabled {
		pc, _, _, _ := runtime.Caller(1)
		funcName := runtime.FuncForPC(pc).Name()
		log.SetOutput(os.Stdout)
		log.Printf("[MSG] [%s] %s", funcName, fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.Enabled {
		pc, _, _, _ := runtime.Caller(1)
		funcName := runtime.FuncForPC(pc).Name()
		log.SetOutput(os.Stderr)
		log.Printf("[ERR] [%s] %s", funcName, fmt.Sprintf(format, args...))
	}
}

func Errorf(format string, args ...interface{}) error {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	return fmt.Errorf("%s: %s", funcName, fmt.Sprintf(format, args...))
}
