package models

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	fmtHeader        = "### [ %s ]"
	fmtComment       = "### %s"
	fmtInfo          = fmtHeader + "\n" + fmtComment
	fmtAppend        = "%s%s"
	fmtAppendNewline = "%s\n%s\n"
)

func (ss *ServerSettings) FromEnv() error {
	return env.Parse(ss)
}

func (ss *ServerSettings) ToEnv() string {
	return ss.String()
}

// String prints the config variables in .env format.
func (ss ServerSettings) String() string {
	str := ss.Runtime.String() + "\n\n"
	str = fmt.Sprintf(fmtAppendNewline, str, ss.Antispam)
	str = fmt.Sprintf(fmtAppendNewline, str, ss.Slack)
	str = fmt.Sprintf(fmtAppendNewline, str, ss.Twitter)

	str = strings.ReplaceAll(str, "\n\n\n", "\n\n")

	return str
}

// FromEnv parses the environment into a ServerSettings struct and yields it, along with an error.
func FromEnv() (*ServerSettings, error) {
	cfg := &ServerSettings{}

	err := cfg.FromEnv()

	return cfg, err
}
