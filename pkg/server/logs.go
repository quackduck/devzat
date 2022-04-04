package server

import (
	"io"
	"log"
	"os"
)

const (
	defaultLogFileName = "log.txt"
)

type serverLogs struct {
	logFile io.WriteCloser
	log     *log.Logger
}

func (sl *serverLogs) init() error {
	f, err := os.OpenFile(defaultLogFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	sl.logFile = f
	sl.log = log.New(io.MultiWriter(sl.logFile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func (s *Server) Log() *log.Logger {
	return s.log
}

func (s *Server) LogFile() io.WriteCloser {
	return s.logFile
}
