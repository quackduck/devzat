package pkg

import (
	"strings"
	"time"
)

func PrintPrettyDuration(d time.Duration) string {
	s := strings.TrimSpace(strings.TrimSuffix(d.Round(time.Minute).String(), "0s"))
	if s == "" { // we cut off the seconds so if there's nothing in the string it means it was made of only seconds.
		s = "< 1m"
	}

	return s
}
