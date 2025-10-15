package main

import (
	"sync"
)

const adminLogSize = 1024

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

func (al *adminLog) formatLogs() []string {
	al.lock.Lock()
	defer al.lock.Unlock()

	if al.isLogFull() {
		return al.formatLogsFull()
	} else {
		return al.formatLogsPartial()
	}
}

func (al *adminLog) isLogFull() bool {
	return al.totalWrittenLines >= adminLogSize
}

func (al *adminLog) formatLogsPartial() []string {
	return al.logLines[0:al.totalWrittenLines]
}

func (al *adminLog) formatLogsFull() []string {
	ret := make([]string, adminLogSize)
	for i := range adminLogSize {
		ret[i] = al.logLines[(al.writeIndex+i)%adminLogSize]
	}
	return ret
}

/* ------------------------------- Public API ------------------------------- */

type AdminLogWriter struct{}

func (_ AdminLogWriter) Write(p []byte) (int, error) {
	globalAdminLog.addLine(p)
	return len(p), nil
}

func GetAdminLog() []string {
	return globalAdminLog.formatLogs()
}
