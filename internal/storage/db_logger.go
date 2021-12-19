package storage

import (
	"log"
)

type Level int

const (
	ERROR Level = iota
	WARNING
	INFO
	DEBUG
)

type DBLogger struct {
	Level
}

func (l *DBLogger) Errorf(f string, v ...interface{}) {
	if l.Level >= ERROR {
		log.Printf(f, v...)
	}
}

func (l *DBLogger) Warningf(f string, v ...interface{}) {
	if l.Level >= WARNING {
		log.Printf(f, v...)
	}
}

func (l *DBLogger) Infof(f string, v ...interface{}) {
	if l.Level >= INFO {
		log.Printf(f, v...)
	}
}

func (l *DBLogger) Debugf(f string, v ...interface{}) {
	if l.Level >= DEBUG {
		log.Printf(f, v...)
	}
}
