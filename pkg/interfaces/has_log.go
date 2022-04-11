package interfaces

import (
	"io"

	"github.com/rs/zerolog"
)

type hasLog interface {
	Log() *zerolog.Logger
	LogFile() io.WriteCloser
}
