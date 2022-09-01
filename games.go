package main

import (
	"fmt"
	"strings"

	"github.com/shurcooL/tictactoe"
)

var (
	tttGame       = new(tictactoe.Board)
	currentPlayer = tictactoe.X
	hangGame      = new(hangman)
)

type hangman struct {
	word      string
	triesLeft int
	guesses   string // string containing all the guessed characters
}

func hangPrint(hangGame *hangman) string {
	display := ""
	for _, c := range hangGame.word {
		if strings.ContainsRune(hangGame.guesses, c) {
			display += string(c)
		} else {
			display += "_"
		}
	}
	return display
}

func tttPrint(cells [9]tictactoe.State) string {
	return strings.ReplaceAll(strings.ReplaceAll(

		fmt.Sprintf(` %v │ %v │ %v 
───┼───┼───
 %v │ %v │ %v 
───┼───┼───
 %v │ %v │ %v `, cells[0], cells[1], cells[2],
			cells[3], cells[4], cells[5],
			cells[6], cells[7], cells[8]),

		tictactoe.X.String(), Chalk.BrightYellow(tictactoe.X.String())), // add some coloring
		tictactoe.O.String(), Chalk.BrightGreen(tictactoe.O.String()))
}
