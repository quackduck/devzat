package hang

import (
	"strconv"
	"strings"
)

const (
	name     = "=<user>"
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

func (c *Command) Fn(rest string, u pkg.User) error {
	devbot := u.Room().Bot().Name()

	if len(rest) > 1 {
		if !u.IsSlack {
			u.Writeln(u.Name, "hang "+rest)
			u.Writeln(devbot, "(that word won't show dw)")
		}

		r.state[u.Room().Name] = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
		u.Room().BotCast(u.Name + " has started a new game of Hangman! Guess letters with hang <letter>")
		u.Room().BotCast("```\n" + hangPrint(r.state[u.Room().Name]) + "\nTries: " + strconv.Itoa(r.state[u.Room().Name].triesLeft) + "\n```")

		return nil
	}

	if !u.IsSlack {
		u.Room().Broadcast(u.Name, "hang "+rest)
	}

	if strings.Trim(r.state[u.Room().Name].word, r.state[u.Room().Name].guesses) == "" {
		u.Room().BotCast("The game has ended. Start a new game with hang <word>")

		return nil
	}

	if len(rest) == 0 {
		u.Room().BotCast("Start a new game with hang <word> or guess with hang <letter>")

		return nil
	}

	if r.state[u.Room().Name].triesLeft == 0 {
		u.Room().BotCast("No more tries! The word was " + r.state[u.Room().Name].word)
		return nil
	}

	if strings.Contains(r.state[u.Room().Name].guesses, rest) {
		u.Room().BotCast("You already guessed " + rest)
		return nil
	}

	r.state[u.Room().Name].guesses += rest
	if !(strings.Contains(r.state[u.Room().Name].word, rest)) {
		r.state[u.Room().Name].triesLeft--
	}

	display := hangPrint(r.state[u.Room().Name])
	u.Room().BotCast("```\n" + display + "\nTries: " + strconv.Itoa(r.state[u.Room().Name].triesLeft) + "\n```")
	if strings.Trim(r.state[u.Room().Name].word, r.state[u.Room().Name].guesses) == "" {
		u.Room().BotCast("You got it! The word was " + r.state[u.Room().Name].word)
	} else if r.state[u.Room().Name].triesLeft == 0 {
		u.Room().BotCast("No more tries! The word was " + r.state[u.Room().Name].word)
	}

	return nil
}
