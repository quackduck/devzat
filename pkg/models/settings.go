package models

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

const (
	configDirName = "devzat"
)

type ServerSettings struct {
	Antispam AntispamSettings
	Slack    SlackSettings
	Twitter  TwitterSettings
	Runtime  runtimeSettings
}

func (ss *ServerSettings) Init() error {
	return ss.FromEnv()
}

func (ss *ServerSettings) ConfigDir() string {
	if ss.Runtime.ConfigDir == "" {
		ss.setDefaultCfgDir()
	}

	return ss.Runtime.ConfigDir
}

func (ss *ServerSettings) SetConfigDir(d string) {
	// handle default case if incoming path is not good
	if _, err := filepath.Abs(d); err != nil || d == "" {
		ss.setDefaultCfgDir()
		return
	}

	ss.Runtime.ConfigDir = d

	ss.ensureConfigDirExists() // make sure it exists
}

func (ss *ServerSettings) ensureConfigDirExists() {
	_ = os.MkdirAll(ss.Runtime.ConfigDir, os.ModePerm)
}

func (ss *ServerSettings) setDefaultCfgDir() {
	cfgDir, _ := os.UserConfigDir()
	ss.Runtime.ConfigDir = filepath.Join(cfgDir, configDirName)
}

func (ss *ServerSettings) ConfigFileName() string        { return ss.Runtime.COnfigFile }
func (ss *ServerSettings) SetConfigFileName(name string) { ss.Runtime.COnfigFile = name }

func (ss *ServerSettings) GetConfigFile() (*os.File, error) {
	cfgFilePath := filepath.Join(ss.Runtime.ConfigDir, ss.Runtime.COnfigFile)

	if _, err := os.Stat(cfgFilePath); os.IsNotExist(err) {
		if errNew := ss.SaveConfigFile(); err != nil {
			return nil, errNew
		}
	} else if err != nil {
		return nil, err
	}

	f, err := os.Open(cfgFilePath)
	if err != nil {
		return nil, err
	}

	return f, err
}

func (ss *ServerSettings) SaveConfigFile() error {
	path := filepath.Join(ss.Runtime.ConfigDir, ss.Runtime.COnfigFile)
	_ = os.MkdirAll(ss.Runtime.ConfigDir, 0777)

	return os.WriteFile(path, []byte(ss.String()), 0777)
}

func dump(cfg interface{}) string {
	c := reflect.ValueOf(cfg)
	t := c.Type()

	fmtStr := "%s\n%s"
	str := ""

	for i := 0; i < c.NumField(); i++ {
		key, found := t.Field(i).Tag.Lookup("env")
		if !found {
			continue
		}

		// TODO handle printing slices with custom separators
		// see hack below
		_, isMulti := t.Field(i).Tag.Lookup("envSeparator")

		def := t.Field(i).Tag.Get("envDefault")

		if c.Field(i).Kind() == reflect.Struct {
			continue
		}

		val := fmt.Sprintf("%v", c.Field(i).Interface())
		fmtLine := "%s=%s"

		// HACK: only prints default for things that are slices
		// we need to format the reflected string slice here
		if val == def || val == "" || val == "0" || val == "0s" || isMulti {
			val = def
			fmtLine = "# %s=%s"
		}

		line := fmt.Sprintf(fmtLine, key, val)
		str = fmt.Sprintf(fmtStr, str, line)
	}

	return str
}
