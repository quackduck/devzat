package hang

import "strings"

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
