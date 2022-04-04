package interfaces

import (
	"io"
	"log"
)

type hasLog interface {
	Log() *log.Logger
	LogFile() io.WriteCloser
}
