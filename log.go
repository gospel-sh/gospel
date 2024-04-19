// Gospel - Golang Simple Extensible Web Framework
// Copyright (C) 2019-2024 - The Gospel Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package gospel

import (
	"io/ioutil"
	"log"
	"os"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

type CustomLogger struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
}

func NewCustomLogger(logLevel LogLevel) *CustomLogger {
	infoWriter := ioutil.Discard
	warningWriter := ioutil.Discard
	errorWriter := ioutil.Discard
	debugWriter := ioutil.Discard

	switch logLevel {
	case DEBUG:
		debugWriter = os.Stdout
		fallthrough
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
		debugLogger:   log.New(debugWriter, "DEBUG:", log.Ldate|log.Ltime),
		infoLogger:    log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime),
		warningLogger: log.New(warningWriter, "WARNING: ", log.Ldate|log.Ltime),
		errorLogger:   log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime),
	}
}

func (cl *CustomLogger) Debug(format string, v ...interface{}) {
	cl.debugLogger.Printf(format, v...)
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
