package ascii_art

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	"math/rand"
)

const (
	name     = "art"
	argsInfo = ""
	info     = "Show some ascii art"
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

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisLow
}

const lolNotImplemented = `
⠀⠀⠀⠀⠀⠀⠚⠛⠔⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠄⠄⠈⢊⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣀⠈⠍⡅⠆⠪⠄⣀⣀⡠⣀⣀⣠⡤⣀⡠⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⢀⠤⣮⠲⣿⠿⠦⣠⣴⣿⠘⣻⣧⣤⡠⠛⠛⢄⠐⠑⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⢠⠾⠵⠧⠟⠬⠭⣵⣶⣿⣻⡿⠠⠭⠝⣻⡥⠔⠪⡙⢠⣶⡕⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠈⢘⠟⣗⣠⠀⠀⠙⠛⠛⠉⠀⠀⡐⢫⣉⡍⠂⠀⣋⢿⣿⠏⡇⠀⠀⠀⠀⢀⠔⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠧⣫⣄⠀⠉⠆⠀⠀⠀⠀⠀⠠⠋⠀⢀⣤⡑⡀⠻⣢⠀⣶⠈⣆⠔⠒⠲⠕⡇⣇⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⢃⠻⠗⠀⢀⠌⡀⡀⠀⠀⠀⠠⡀⠀⠘⠿⡣⠁⠀⡇⠓⠙⣁⠃⠀⠀⡠⠃⢡⣞⢳⡑⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠁⠲⣒⠥⠊⡸⠀⠀⠀⠀⠀⠘⠋⠭⠑⠊⠆⡔⠘⢢⢹⣿⡤⠔⣾⣿⢿⡘⠖⣱⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⢏⣀⣠⠔⠁⡆⠀⠀⢀⣠⣤⣤⣤⠀⠀⠀⠀⠀⠸⢌⡍⠚⣾⢌⡇⠈⣿⣿⣿⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⡇⠀⠀⠀⡠⣾⠿⠛⣋⣉⠭⠄⠀⢴⣿⠿⠆⠀⠸⣀⡒⢪⢌⠀⢹⠿⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⢇⠀⢀⠊⠭⠔⠚⠭⢀⡀⠀⠀⠀⠀⠀⠀⠘⠛⠀⠀⣘⣂⠾⠂⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠢⠌⠀⠀⠀⠀⠰⡿⣿⣿⡏⠉⠉⠉⠉⠉⠩⣿⢻⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣮⣮⢄⠁⠀⠀⠀⠀⠀⠀⠀⡌⣆⠀⢀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣷⣿⠀⠀⠀⠀⠀⠀⠀⠈⠒⠨⣽⣿⠆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⡿⠋⠀⠀⠀⠀⠀⠀⠀⠀
`

func (c *Command) Fn(_ string, u interfaces.User) error {
	art := []string{
		lolNotImplemented,
	}

	u.Room().Broadcast("", art[rand.Intn(len(art))])

	return nil
}
