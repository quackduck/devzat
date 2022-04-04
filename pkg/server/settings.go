package server

import (
	"devzat/pkg"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type serverSettings struct {
	Antispam    antispamSettings `json:"antispam"`
	Slack       slackSettings    `json:"slack"`
	Twitter     twitterSettings  `json:"twitter"`
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

type antispamSettings struct {
	Window      time.Duration `json:"window"`
	LimitWarn   int           `json:"limitWarn"`
	LimitBan    int           `json:"limitBan"`
	BanDuration time.Duration `json:"BanDuration"`
}

type slackSettings struct {
	Offline bool `json:"offline"`
}

type twitterSettings struct {
	Offline bool `json:"offline"`
}

func (ss *serverSettings) init() error {
	// should this instance run offline? (should it not connect to slack or twitter?)
	ss.Slack.Offline = os.Getenv(pkg.EnvOfflineSlack) != ""
	ss.Twitter.Offline = os.Getenv(pkg.EnvOfflineTwitter) != ""

	return nil
}

func (ss *serverSettings) ConfigDir() string     { return ss.dir }
func (ss *serverSettings) SetConfigDir(d string) { ss.dir = d }

func (ss *serverSettings) ConfigFileName() string        { return ss.cfgFileName }
func (ss *serverSettings) SetConfigFileName(name string) { ss.cfgFileName = name }

func (ss *serverSettings) GetConfigFile() (*os.File, error) {
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

func (ss *serverSettings) SaveConfigFile() error {
	path := filepath.Join(ss.dir, ss.cfgFileName)
	_ = os.MkdirAll(ss.dir, 0777)

	data, err := json.MarshalIndent(ss, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config file data: %v", err)
	}

	return os.WriteFile(path, data, 0777)
}
