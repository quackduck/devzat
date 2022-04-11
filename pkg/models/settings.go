package models

import (
	"devzat/pkg"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	configDirName = "devzat"
)

type ServerSettings struct {
	Antispam    AntispamSettings `json:"antispam"`
	Slack       SlackSettings    `json:"slack"`
	Twitter     TwitterSettings  `json:"twitter"`
	Port        int              `json:"port"`
	ProfilePort int              `json:"profilePort"`
	Scrollback  int              `json:"scrollbackHistory"`
	dir         string
	cfgFileName string
}

const (
	antispamDefaultLimitWarn   = 30
	antispamDefaultLimitBan    = 50
	antispamDefaultWindow      = time.Second * 5
	antispamDefaultBanDuration = time.Minute * 5
)

type AntispamSettings struct {
	Window      time.Duration `json:"window"`
	LimitWarn   int           `json:"limitWarn"`
	LimitBan    int           `json:"limitBan"`
	BanDuration time.Duration `json:"BanDuration"`
}

type SlackSettings struct {
	Offline bool `json:"offline"`
}

type TwitterSettings struct {
	Offline bool `json:"offline"`
}

func (ss *ServerSettings) Init() error {
	// should this instance run offline? (should it not connect to slack or twitter?)
	ss.Slack.Offline = os.Getenv(pkg.EnvOfflineSlack) != ""
	ss.Twitter.Offline = os.Getenv(pkg.EnvOfflineTwitter) != ""

	return nil
}

func (ss *ServerSettings) ConfigDir() string {
	if ss.dir == "" {
		ss.setDefaultCfgDir()
	}

	return ss.dir
}

func (ss *ServerSettings) SetConfigDir(d string) {
	if _, err := filepath.Abs(d); err == nil || d == "" {
		ss.setDefaultCfgDir()
		return
	}

	ss.dir = d
}

func (ss *ServerSettings) setDefaultCfgDir() {
	cfgDir, _ := os.UserConfigDir()
	ss.dir = filepath.Join(cfgDir, configDirName)
}

func (ss *ServerSettings) ConfigFileName() string        { return ss.cfgFileName }
func (ss *ServerSettings) SetConfigFileName(name string) { ss.cfgFileName = name }

func (ss *ServerSettings) GetConfigFile() (*os.File, error) {
	cfgFilePath := filepath.Join(ss.dir, ss.cfgFileName)

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
	path := filepath.Join(ss.dir, ss.cfgFileName)
	_ = os.MkdirAll(ss.dir, 0777)

	data, err := json.MarshalIndent(ss, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config file data: %v", err)
	}

	return os.WriteFile(path, data, 0777)
}
