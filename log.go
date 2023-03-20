package gospel

import (
	"io/ioutil"
	"log"
	"os"
)

type LogLevel int

const (
	INFO LogLevel = iota
	WARNING
	ERROR
)

type CustomLogger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
}

func NewCustomLogger(logLevel LogLevel) *CustomLogger {
	infoWriter := ioutil.Discard
	warningWriter := ioutil.Discard
	errorWriter := ioutil.Discard

	switch logLevel {
	case INFO:
		infoWriter = os.Stdout
		fallthrough
	case WARNING:
		warningWriter = os.Stdout
		fallthrough
	case ERROR:
		errorWriter = os.Stderr
	}

	return &CustomLogger{
		infoLogger:    log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime),
		warningLogger: log.New(warningWriter, "WARNING: ", log.Ldate|log.Ltime),
		errorLogger:   log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime),
	}
}

func (cl *CustomLogger) Info(format string, v ...interface{}) {
	cl.infoLogger.Printf(format, v...)
}

func (cl *CustomLogger) Warning(format string, v ...interface{}) {
	cl.warningLogger.Printf(format, v...)
}

func (cl *CustomLogger) Error(format string, v ...interface{}) {
	cl.errorLogger.Printf(format, v...)
}

var Log = NewCustomLogger(INFO)
