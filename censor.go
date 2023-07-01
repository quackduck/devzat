package main

import (
	"fmt"
	"strings"

	"github.com/CaenJones/goaway"
)

func main() {
	text := "My stupid iphone broke on the last stupid day before retirement."

	// Split the text into words
	words := splitTextIntoWords(text)

	// Filter out blocked words
	filteredWords := goaway.FilterBlockedWords(words)

	// Join the filtered words back into a string
	filteredText := joinWordsIntoText(filteredWords)

	fmt.Println(filteredText)
}

// splitTextIntoWords splits the given text into a slice of words
func splitTextIntoWords(text string) []string {
	// Split the text by whitespace
	words := strings.Fields(text)
	return words
}

// joinWordsIntoText joins the given slice of words into a string
func joinWordsIntoText(words []string) string {
	// Join the words with a space separator
	text := strings.Join(words, " ")
	return text
}
