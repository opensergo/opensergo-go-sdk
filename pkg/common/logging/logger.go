// Copyright 2022, OpenSergo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

// Logger the common interface for logging.
type Logger interface {

	// Print logs message no format as what the msg presents.
	Print(msg string)

	// DebugEnabled judge is the DebugLevel enabled
	DebugEnabled() bool
	// Debug logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Debug(msg string, keysAndValues ...interface{})

	// InfoEnabled judge is the InfoLevel enabled
	InfoEnabled() bool
	// Info logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Info(msg string, keysAndValues ...interface{})

	// WarnEnabled judge is the WarnLevel enabled
	WarnEnabled() bool
	// Warn logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Warn(msg string, keysAndValues ...interface{})

	// ErrorEnabled judge is the ErrorLevel enabled
	ErrorEnabled() bool
	// Error logs an error message with error and the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Error(err error, msg string, keysAndValues ...interface{})
}

// Print logs message no format as what the msg presents.
func Print(msg string) {
	doLog("Print", nil, msg)
}

// Debug logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func Debug(msg string, keysAndValues ...interface{}) {
	doLog("Debug", nil, msg, keysAndValues...)
}

// DebugWithCallerDepth logs a non-error message with the given key/value pairs as context.
//
// logCallerDepth: to calculate the caller:line
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func DebugWithCallerDepth(logger Logger, logFormat LogFormat, logCallerDepth int, msg string, keysAndValues ...interface{}) {
	if !logger.DebugEnabled() {
		return
	}
	logger.Print(AssembleMsg(logFormat, logCallerDepth, "DEBUG", msg, nil, false, keysAndValues...))
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func Info(msg string, keysAndValues ...interface{}) {
	doLog("Info", nil, msg, keysAndValues...)
}

// InfoWithCallerDepth logs a non-error message with the given key/value pairs as context.
//
// logCallerDepth: to calculate the caller:line
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func InfoWithCallerDepth(logger Logger, logFormat LogFormat, logCallerDepth int, msg string, keysAndValues ...interface{}) {
	if !logger.InfoEnabled() {
		return
	}
	logger.Print(AssembleMsg(logFormat, logCallerDepth, "INFO", msg, nil, false, keysAndValues...))
}

// Warn logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func Warn(msg string, keysAndValues ...interface{}) {
	doLog("Warn", nil, msg, keysAndValues...)
}

// WarnWithCallerDepth logs a non-error message with the given key/value pairs as context.
//
// logCallerDepth: to calculate the caller:line
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func WarnWithCallerDepth(logger Logger, logFormat LogFormat, logCallerDepth int, msg string, keysAndValues ...interface{}) {
	if !logger.WarnEnabled() {
		return
	}

	logger.Print(AssembleMsg(logFormat, logCallerDepth, "WARN", msg, nil, false, keysAndValues...))
}

// Error logs an error message with error and the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func Error(err error, msg string, keysAndValues ...interface{}) {
	doLog("Error", err, msg, keysAndValues...)
}

// ErrorWithCallerDepth logs an error message with error and the given key/value pairs as context.
//
// logCallerDepth: to calculate the caller:line
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func ErrorWithCallerDepth(logger Logger, logFormat LogFormat, logCallerDepth int, err error, errorWithStack bool, msg string, keysAndValues ...interface{}) {
	if !logger.ErrorEnabled() {
		return
	}
	logger.Print(AssembleMsg(logFormat, logCallerDepth, "ERROR", msg, err, errorWithStack, keysAndValues...))
}

var (
	loggers      = make([]Logger, 0)
	loggersMutex sync.Mutex
)

// AddLogger add the Logger into loggerSlice
func AddLogger(loggerAppend Logger) int {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	loggers = append(loggers, loggerAppend)
	return len(loggers)
}

// ClearLogger clear the loggerSlice
func ClearLogger() {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()

	loggers = make([]Logger, 0)
}

// doLog do log
// funcNameFromInterface funcName in Logger
// err
// msg
// keysAndValues
func doLog(funcNameFromInterface string, err error, msg string, keysAndValues ...interface{}) {
	if len(loggers) == 0 {
		defaultConsoleLogger := NewDefaultConsoleLogger(InfoLevel)
		AddLogger(defaultConsoleLogger)
	}

	if len(loggers) > 0 {
		for _, logger := range loggers {
			invokeLogger(logger, funcNameFromInterface, err, msg, keysAndValues...)
		}
	}
}

// invokeLogger do log actually by invoke function of Logger
// logger Logger to print
// funcNameFromInterface funcName in Logger
// err
// msg
// keysAndValues
func invokeLogger(logger Logger, funcNameFromInterface string, err error, msg string, keysAndValues ...interface{}) {
	method, ok := reflect.TypeOf(logger).MethodByName(funcNameFromInterface)
	if !ok {
		assembleMsg := AssembleMsg(SeparateFormat, 4, "WARN", "no function named '"+funcNameFromInterface+"' was found in interface 'opensergo-go/pkg/logging/Logger'", nil, false)
		logger.Print(assembleMsg)
		return
	}

	keysAndValuesLen := len(keysAndValues)
	params := make([]reflect.Value, 0)
	params = append(params, reflect.ValueOf(logger))
	if "Error" == funcNameFromInterface {
		if err == nil {
			err = errors.New("")
		}
		params = append(params, reflect.ValueOf(err))
	}
	params = append(params, reflect.ValueOf(msg))

	if keysAndValuesLen != 0 {
		if keysAndValuesLen == 1 && keysAndValues[0] == nil {

		} else {
			for _, keyOrValue := range keysAndValues {
				params = append(params, reflect.ValueOf(keyOrValue))
			}
		}
	}
	method.Func.Call(params)
}
