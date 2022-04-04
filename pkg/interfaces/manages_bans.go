package interfaces

import (
	"devzat/pkg/models"
	"time"
)

type managesBans interface {
	SaveBans() error
	ReadBans() error
	BansContains(addr string, id string) bool
	BanUser(strBanner string, victim User)
	GetBanList() []models.Ban
	BanUserForDuration(banner string, victim User, dur time.Duration)
	UnbanUser(toUnban string) error
}
