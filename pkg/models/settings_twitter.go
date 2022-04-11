package models

import "fmt"

type TwitterSettings struct {
	Offline bool `env:"DEVZAT_TWITTER_OFFLINE" envDefault:"true"`
}

func (cfg TwitterSettings) String() string {
	const (
		header  = "Twitter"
		comment = "settings for twitter integration"
	)

	info := fmt.Sprintf(fmtInfo, header, comment)

	return fmt.Sprintf(fmtAppend, info, dump(cfg))
}
