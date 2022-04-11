package server

import (
	"github.com/rs/zerolog"
	"io"
	"os"
)

const (
	defaultLogFileName = "log.txt"
)

func (sl *Server) initLogs() error {
	f, err := os.OpenFile(defaultLogFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	sl.Logger = zerolog.New(f).With().Logger()

	sl.logFile = f

	return nil
}

func (s *Server) LogFile() io.WriteCloser {
	return s.logFile
}

func (s *Server) Log() *zerolog.Logger {
	return &s.Logger
}
