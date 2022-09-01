package main

import (
	"encoding/base64"
	goaway "github.com/TwiN/go-away"
)

var detector = goaway.NewProfanityDetector().WithSanitizeSpaces(false)

func rmBadWords(text string) string {
	if !Config.Censor {
		return text
	}
	return detector.Censor(text)
}

func init() {
	okayIshWords := []string{"ZnVjaw==", "Y3JhcA==", "c2hpdA==", "YXJzZQ==", "YXNz", "YnV0dA==", "cGlzcw=="} // base 64 encoded okay-ish swears
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
}
