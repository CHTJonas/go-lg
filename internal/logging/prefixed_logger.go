package logging

import (
	"log"
	"os"
)

type Level int

const (
	ERROR Level = iota
	WARNING
	INFO
	DEBUG
)

type PrefixedLogger struct {
	*log.Logger
	Level
}

func NewPrefixedLogger(prefix string, level Level) *PrefixedLogger {
	str := "[" + prefix + "] "
	return &PrefixedLogger{
		Logger: log.New(os.Stderr, str, log.Lmsgprefix),
		Level:  level,
	}
}

func (l *PrefixedLogger) Errorf(f string, v ...interface{}) {
	if l.Level >= ERROR {
		l.Printf("ERROR: "+f, v...)
	}
}

func (l *PrefixedLogger) Warningf(f string, v ...interface{}) {
	if l.Level >= WARNING {
		l.Printf("WARNING: "+f, v...)
	}
}

func (l *PrefixedLogger) Infof(f string, v ...interface{}) {
	if l.Level >= INFO {
		l.Printf("INFO: "+f, v...)
	}
}

func (l *PrefixedLogger) Debugf(f string, v ...interface{}) {
	if l.Level >= DEBUG {
		l.Printf("DEBUG: "+f, v...)
	}
}
