package main

import (
	"sync"
)

const adminLogSize = 4

// This struct contains copied log lines meant to be read by admins.
// They are stored as a circular buffer for ease of write.
type adminLog struct {
	logLines         [adminLogSize]string
	writeIndex       int
	totalWrittenLines int
	lock             sync.Mutex
}

var globalAdminLog = adminLog{totalWrittenLines: 0, writeIndex: 0}

func (al *adminLog) addLine(p []byte) {
	al.lock.Lock()
	defer al.lock.Unlock()

	al.logLines[al.writeIndex%adminLogSize] = string(p)
	al.writeIndex++
	al.totalWrittenLines++
}

type AdminLogWriter struct{}

func (_ AdminLogWriter) Write(p []byte) (int, error) {
	globalAdminLog.addLine(p)
	return len(p), nil
}
