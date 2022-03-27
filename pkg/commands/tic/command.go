package tic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jwalton/gchalk"
	"github.com/shurcooL/tictactoe"

	"devzat/pkg/user"
)

const (
	name     = "tic"
	argsInfo = ""
	info     = ""
)

type state = struct {
	Board         *tictactoe.Board
	currentPlayer tictactoe.State
}

type Command struct {
	state map[string]*state // each room has its own state
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

func (c *Command) IsRest() bool {
	return false
}

func (c *Command) IsSecret() bool {
	return false
}

func (c *Command) Fn(rest string, u *user.User) error {
	devbot := u.Room.Bot.Name()

	if c.state == nil {
		c.state = make(map[string]*state)
	}

	s := c.getState(u)

	if rest == "" {
		u.Room.Broadcast(devbot, "Starting a new game of Tic Tac Toe! The first player is always X.")
		u.Room.Broadcast(devbot, "Play using tic <cell num>")
		s.currentPlayer = tictactoe.X
		s.Board = new(tictactoe.Board)
		u.Room.Broadcast(devbot, "```\n"+" 1 │ 2 │ 3\n───┼───┼───\n 4 │ 5 │ 6\n───┼───┼───\n 7 │ 8 │ 9\n"+"\n```")
		return nil
	}

	m, err := strconv.Atoi(rest)
	if err != nil {
		u.Room.Broadcast(devbot, "Make sure you're using a number lol")
		return nil
	}

	if m < 1 || m > 9 {
		u.Room.Broadcast(devbot, "Moves are numbers between 1 and 9!")
		return nil
	}

	err = s.Board.Apply(tictactoe.Move(m-1), s.currentPlayer)

	if err != nil {
		u.Room.Broadcast(devbot, err.Error())
		return nil
	}

	u.Room.Broadcast(devbot, "```\n"+tttPrint(s.Board.Cells)+"\n```")
	if s.currentPlayer == tictactoe.X {
		s.currentPlayer = tictactoe.O
	} else {
		s.currentPlayer = tictactoe.X
	}

	if !(s.Board.Condition() == tictactoe.NotEnd) {
		u.Room.Broadcast(devbot, s.Board.Condition().String())
		s.currentPlayer = tictactoe.X
		s.Board = new(tictactoe.Board)
	}

	return nil
}

func (c *Command) getState(u *user.User) *state {
	if c.state[u.Room.Name] == nil {
		c.state[u.Room.Name] = &state{}
	}

	return c.state[u.Room.Name]
}

func tttPrint(cells [9]tictactoe.State) string {
	chalk := gchalk.New(gchalk.ForceLevel(gchalk.LevelAnsi256))
	return strings.ReplaceAll(strings.ReplaceAll(

		fmt.Sprintf(` %v │ %v │ %v 
───┼───┼───
 %v │ %v │ %v 
───┼───┼───
 %v │ %v │ %v `, cells[0], cells[1], cells[2],
			cells[3], cells[4], cells[5],
			cells[6], cells[7], cells[8]),

		tictactoe.X.String(), chalk.BrightYellow(tictactoe.X.String())), // add some coloring
		tictactoe.O.String(), chalk.BrightGreen(tictactoe.O.String()))
}
