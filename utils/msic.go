package utils

import (
	"runtime/debug"
)

var log ILogger

func WithLogger(logger ILogger) {
	log = logger
}

func ErrorPrint() {
	if err := recover(); err != nil {
		log.Error(debug.Stack())
	}
}
