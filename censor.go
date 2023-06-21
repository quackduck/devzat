package main

import (
	"fmt"
	"goaway"
)

func main() {
	text := "This is a sample text with some words that may trigger the profanity filter."

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
	words := make([]string, 0)
	currentWord := ""
	for _, char := range text {
		if char == ' ' {
			// Add the current word to the slice
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			// Append the character to the current word
			currentWord += string(char)
		}
	}

	// Add the last word to the slice (if any)
	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// joinWordsIntoText joins the given slice of words into a string
func joinWordsIntoText(words []string) string {
	// Join the words with a space separator
	text := ""
	for i, word := range words {
		if i > 0 {
			text += " "
		}
		text += word
	}
	return text
}
