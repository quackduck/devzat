package hang

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"strconv"
	"strings"
)

const (
	name     = "hang"
	argsInfo = "<char|word>"
	info     = "start a game of hangman"
)

type Command struct {
	state map[string]*hangman
}

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
	return models.CommandVisNormal
}

func (c *Command) Fn(rest string, u i.User) error {
	devbot := u.Room().Bot().Name()

	if _, alreadyHasGame := c.state[u.Room().Name()]; alreadyHasGame {
		u.Room().BotCast("there's already a game for this room, wait until it's finished.")
		return nil
	}

	if len(rest) > 1 {
		if !u.IsSlack() {
			u.Writeln(u.Name(), "hang "+rest)
			u.Writeln(devbot, "(that word won't show dw)")
		}

		c.state[u.Room().Name()] = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
		u.Room().BotCast(u.Name() + " has started a new game of Hangman! Guess letters with hang <letter>")
		u.Room().BotCast("```\n" + hangPrint(c.state[u.Room().Name()]) + "\nTries: " + strconv.Itoa(c.state[u.Room().Name()].triesLeft) + "\n```")

		return nil
	}

	if !u.IsSlack() {
		u.Room().Broadcast(u.Name(), "hang "+rest)
	}

	if strings.Trim(c.state[u.Room().Name()].word, c.state[u.Room().Name()].guesses) == "" {
		u.Room().BotCast("The game has ended. Start a new game with hang <word>")

		return nil
	}

	if len(rest) == 0 {
		u.Room().BotCast("Start a new game with hang <word> or guess with hang <letter>")

		return nil
	}

	if c.state[u.Room().Name()].triesLeft == 0 {
		u.Room().BotCast("No more tries! The word was " + c.state[u.Room().Name()].word)
		return nil
	}

	if strings.Contains(c.state[u.Room().Name()].guesses, rest) {
		u.Room().BotCast("You already guessed " + rest)
		return nil
	}

	c.state[u.Room().Name()].guesses += rest
	if !(strings.Contains(c.state[u.Room().Name()].word, rest)) {
		c.state[u.Room().Name()].triesLeft--
	}

	display := hangPrint(c.state[u.Room().Name()])
	u.Room().BotCast("```\n" + display + "\nTries: " + strconv.Itoa(c.state[u.Room().Name()].triesLeft) + "\n```")
	if strings.Trim(c.state[u.Room().Name()].word, c.state[u.Room().Name()].guesses) == "" {
		u.Room().BotCast("You got it! The word was " + c.state[u.Room().Name()].word)
	} else if c.state[u.Room().Name()].triesLeft == 0 {
		u.Room().BotCast("No more tries! The word was " + c.state[u.Room().Name()].word)
	}

	return nil
}
