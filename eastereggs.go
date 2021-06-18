package main

import (
	"fmt"
	"strings"
)

func triggerEasterEggs(u *user, message string) {
	if strings.Contains(message, "rm -rf /") {
		broadcast(u, "rm -rf /: Permission denied, you troll")
		return
	}
	if strings.Contains(message, "ls") {
		result := ""
		for _, command := range commands {
			result = fmt.Sprintf("%s *%s", result, command.name)
		}
		broadcast(u, result)
		return
	}
	if strings.Contains(message, "easter") {
		broadcast(u, "eggs")
		return
	}
	if strings.Contains(message, "cat README.md") {
		helpCommand(u, nil)
		return
	}
	if strings.Contains(message, "cat ") {
		broadcast(u, fmt.Sprintf("cat: Permission denied"))
		return
	}

}
