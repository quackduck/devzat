package unban

import (
	"devzat/pkg/user"
	"fmt"
)

const (
	name     = ""
	argsInfo = ""
	info     = ""
)

type Command struct{}

func (c *Command) Name() string {
	return name
}

func (c *Command) ArgsInfo() string {
	return argsInfo
}

func (c *Command) Info() string {
	return info
}

func (c *Command) IsRest() bool {
	return false
}

func (c *Command) IsSecret() bool {
	return false
}

func (c *Command) Fn(toUnban string, u *user.User) error {
	isAdmin, errCheckAdmin := checkIsAdmin(u)
	if errCheckAdmin != nil {
		return fmt.Errorf("could not unban: %v", errCheckAdmin)
	}

	if !isAdmin {
		u.Room.Broadcast(devbot, "Not authorized")
		return nil
	}

	if unbanIDorIP(toUnban) {
		u.Room.Broadcast(devbot, "Unbanned person: "+toUnban)
		saveBans()

		return nil
	}

	u.Room.Broadcast(devbot, "I couldn't find that person")

	return nil
}

// unbanIDorIP unbans an ID or an IP, but does NOT save bans to the bans file.
// It returns whether the person was found, and so, whether the bans slice was modified.
func unbanIDorIP(toUnban string) bool {
	for i := 0; i < len(bans); i++ {
		if bans[i].ID == toUnban || bans[i].Addr == toUnban { // allow unbanning by either ID or IP
			// remove this ban
			bans = append(bans[:i], bans[i+1:]...)
			return true
		}
	}
	return false
}
