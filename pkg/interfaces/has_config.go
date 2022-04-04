package interfaces

import "os"

type hasConfig interface {
	SetConfigDir(string)
	ConfigDir() string
	ConfigFileName() string
	SetConfigFileName(string)
	GetConfigFile() (*os.File, error)
	SaveConfigFile() error
}
