package models

import (
	"fmt"
	"time"
)

type AntispamSettings struct {
	Window      time.Duration `env:"DEVZAT_SPAM_WINDOW" envDefault:"5s"`
	LimitWarn   int           `env:"DEVZAT_SPAM_LIMITWARN" envDefault:"30"`
	LimitBan    int           `env:"DEVZAT_SPAM_LIMITBAN" envDefault:"50"`
	BanDuration time.Duration `env:"DEVZAT_SPAM_BANDURATION" envDefault:"5m0s"`
}

func (cfg AntispamSettings) String() string {
	const (
		header  = "Anti-spam"
		comment = "settings for controlling spam"
	)

	info := fmt.Sprintf(fmtInfo, header, comment)

	return fmt.Sprintf(fmtAppend, info, dump(cfg))
}
