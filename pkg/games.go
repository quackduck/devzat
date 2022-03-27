package pkg

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

		tictactoe.X.String(), chalk.BrightYellow(tictactoe.X.String())), // add some coloring
		tictactoe.O.String(), chalk.BrightGreen(tictactoe.O.String()))
}
