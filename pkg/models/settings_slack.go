package models

import "fmt"

type SlackSettings struct {
	Offline bool `env:"DEVZAT_SLACK_OFFLINE" envDefault:"true"`
}

func (cfg SlackSettings) String() string {
	const (
		header  = "Slack"
		comment = "settings for slack integration"
	)

	info := fmt.Sprintf(fmtInfo, header, comment)

	return fmt.Sprintf(fmtAppend, info, dump(cfg))
}
