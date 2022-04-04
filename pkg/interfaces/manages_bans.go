package interfaces

import "time"

type managesBans interface {
	SaveBans() error
	ReadBans() error
	BansContains(addr string, id string) bool
	BanUser(strBanner string, victim User)
	BanUserForDuration(banner string, victim User, dur time.Duration)
	UnbanUser(toUnban string) error
}
