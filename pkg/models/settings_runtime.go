package models

import "fmt"

type runtimeSettings struct {
	Port        int    `env:"DEVZAT_PORT" envDefault:"22"`
	ProfilePort int    `env:"DEVZAT_PROFILEPORT" envDefault:"5555"`
	Scrollback  int    `env:"DEVZAT_SCROLLBACKHISTORY" envDefault:"16"`
	ConfigDir   string `env:"DEVZAT_CFGDIR"`
	COnfigFile  string `env:"DEVZAT_CFGFILENAME"`
}

func (cfg runtimeSettings) String() string {
	const (
		header  = "Runtime"
		comment = "settings for the devzat runtime"
	)

	info := fmt.Sprintf(fmtInfo, header, comment)

	return fmt.Sprintf(fmtAppend, info, dump(cfg))
}
