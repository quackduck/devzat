package v2

import (
	"devzat/pkg"
	"os"
	"time"
)

type settings struct {
	Antispam antispamSettings `json:"antispam"`
	Slack    slackSettings    `json:"slack"`
	Twitter  twitterSettings  `json:"twitter"`
}

func (s settings) init() error {
	// should this instance run offline? (should it not connect to slack or twitter?)
	s.Slack.Offline = os.Getenv(pkg.EnvOfflineSlack) != ""
	s.Twitter.Offline = os.Getenv(pkg.EnvOfflineTwitter) != ""

	return nil
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
