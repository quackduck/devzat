package server

import (
	"strings"

	"github.com/acarl005/stripansi"

	"devzat/pkg/interfaces"
)

func (s *Server) autocompleteCallback(u interfaces.User, line string, pos int, key rune) (string, int, bool) {
	if key != '\t' {
		return "", pos, false
	}

	// Autocomplete a username

	// Split the input string to look for @<name>
	words := strings.Fields(line)

	toAdd := s.userMentionAutocomplete(u, words)
	if toAdd != "" {
		return line + toAdd, pos + len(toAdd), true
	}

	toAdd = s.roomAutocomplete(words)
	if toAdd != "" {
		return line + toAdd, pos + len(toAdd), true
	}

	//return line + toAdd + " ", pos + len(toAdd) + 1, true
	return "", pos, false
}

func (s *Server) userMentionAutocomplete(u interfaces.User, words []string) string {
	if len(words) < 1 {
		return ""
	}

	// Check the last word and see if it's trying to refer to a User
	if words[len(words)-1][0] == '@' || (len(words)-1 == 0 && words[0][0] == '=') { // mentioning someone or dm-ing someone
		inputWord := words[len(words)-1][1:] // slice the @ or = off

		for _, user := range u.Room().AllUsers() {
			strippedName := stripansi.Strip(user.Name())
			toAdd := strings.TrimPrefix(strippedName, inputWord)
			if toAdd != strippedName { // there was a match, and some text got trimmed!
				return toAdd + " "
			}
		}
	}

	return ""
}

func (s *Server) roomAutocomplete(words []string) string {
	// trying to refer to a room?
	if len(words) > 0 && words[len(words)-1][0] == '#' {
		// don't slice the # off, since the room name includes it
		for name := range s.rooms {
			toAdd := strings.TrimPrefix(name, words[len(words)-1])
			if toAdd != name { // there was a match, and some text got trimmed!
				return toAdd + " "
			}
		}
	}

	return ""
}
