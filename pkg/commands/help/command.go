package help

import (
	"devzat/pkg/user"
)

const defaultHelpFileName = "help.txt"

const defaultHelpMessage = `Welcome to Devzat!
Devzat is chat over SSH: github.com/quackduck/devzat  
Because there's SSH apps on all platforms, even on mobile, you can join from anywhere.

Run cmds to see a list of commands.

Interesting features:
* Rooms! Run cd to see all Rooms and use cd #foo to join a new room.
* Markdown support! Tables, headers, italics and everything. Just use \\n in place of newlines.
* Code syntax highlighting. Use Markdown fences to send code. Run eg-code to see an example.
* Direct messages! Send a quick DirectMessage using =User <msg> or stay in DMs by running cd @User.
* Timezone support, use tz Continent/City to set your timezone.
* Built in Tic Tac Toe and Hangman! Run tic or hang <word> to start new games.
* Emoji replacements! \:rocket\: => :rocket: (like on Slack and Discord)

For replacing newlines, I often use bulkseotools.com/add-remove-line-breaks.php.

Join the Devzat discord server: https://discord.gg/5AUjJvBHeT

Made by Ishan Goel with feature ideas from friends.  
Thanks to Caleb Denio for lending his server!`

const (
	name     = "help"
	argsInfo = "<msg>"
	info     = "DirectMessage <User> with <msg>"
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

func (c *Command) Fn(_ string, u *user.User) error {
	u.Room.Broadcast("", defaultHelpMessage)

	return nil
}
