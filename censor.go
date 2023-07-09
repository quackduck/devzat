package main

import (
	"encoding/base64"
	"fmt"

	goaway "github.com/CaenJones/go-away"
)

var detector *goaway.ProfanityDetector

func rmBadWords(text string) string {
	if !Config.Censor {
		return text
	}
	return detector.Censor(text)
}

func init() {
	okayIshWords := []string{"ZnVjaw==", "Y3JhcA==", "c2hpdA==", "YXJzZQ==", "YXNz", "YnV0dA==", "cGlzcw=="} // base64 encoded okay-ish swears
	for i := 0; i < len(goaway.DefaultProfanities); i++ {
		for _, okayIshWord := range okayIshWords {
			okayIshWordb, _ := base64.StdEncoding.DecodeString(okayIshWord)
			if goaway.DefaultProfanities[i] == string(okayIshWordb) {
				goaway.DefaultProfanities = append(goaway.DefaultProfanities[:i], goaway.DefaultProfanities[i+1:]...)
				i-- // so we don't skip the next word
				break
			}
		}
	}

	detector = goaway.NewProfanityDetector().WithSanitizeSpaces(false).WithFalsePositivesFunc(DefaultFalsePositives)
}

// DefaultFalsePositives is a function that returns true for any word
func DefaultFalsePositives(word string) bool {
	return true
}

func main() {
	text := "This is a test sentence with a bad word."
	censoredText := rmBadWords(text)
	fmt.Println(censoredText)
}
