package logging

import (
	"log"
	"os"
)

type PrefixedLogger struct {
	*log.Logger
}

func NewPrefixedLogger(prefix string) *PrefixedLogger {
	str := "[" + prefix + "] "
	return &PrefixedLogger{
		Logger: log.New(os.Stderr, str, log.Lmsgprefix),
	}
}

func (l *PrefixedLogger) Errorf(f string, v ...interface{}) {
	l.Printf("ERROR: "+f, v...)
}

func (l *PrefixedLogger) Warningf(f string, v ...interface{}) {
	l.Printf("WARNING: "+f, v...)
}

func (l *PrefixedLogger) Infof(f string, v ...interface{}) {
	l.Printf("INFO: "+f, v...)
}

func (l *PrefixedLogger) Debugf(f string, v ...interface{}) {
	l.Printf("DEBUG: "+f, v...)
}
