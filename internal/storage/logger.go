package storage

import (
	"log"
	"os"
)

type storageLog struct {
	*log.Logger
}

func newStorageLogger() *storageLog {
	return &storageLog{
		Logger: log.New(os.Stderr, "[db] ", log.Lmsgprefix),
	}
}

func (l *storageLog) Errorf(f string, v ...interface{}) {
	l.Printf("ERROR: "+f, v...)
}

func (l *storageLog) Warningf(f string, v ...interface{}) {
	l.Printf("WARNING: "+f, v...)
}

func (l *storageLog) Infof(f string, v ...interface{}) {
	l.Printf("INFO: "+f, v...)
}

func (l *storageLog) Debugf(f string, v ...interface{}) {
	l.Printf("DEBUG: "+f, v...)
}
