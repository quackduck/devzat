package server

import (
	"net/http"
	"strings"
)

func (s *Server) ReplaceSlackEmoji(input string) string {
	if len(input) < 4 {
		return input
	}
	emojiName := ""
	result := make([]byte, 0, len(input))
	inEmojiName := false
	for i := 0; i < len(input)-1; i++ {
		if inEmojiName {
			emojiName += string(input[i]) // end result: if input contains "::lol::", emojiName will contain ":lol:". "::lol:: ::cat::" => ":lol::cat:"
		}
		if input[i] == ':' && input[i+1] == ':' {
			inEmojiName = !inEmojiName
		}
		//if !inEmojiName {
		result = append(result, input[i])
		//}
	}
	result = append(result, input[len(input)-1])
	if emojiName != "" {
		toAdd := s.fetchEmoji(strings.Split(strings.ReplaceAll(emojiName[1:len(emojiName)-1], "::", ":"), ":")) // cut the ':' at the start and end

		result = append(result, toAdd...)
	}
	return string(result)
}

// accepts a ':' separated list of emoji
func (s *Server) fetchEmoji(names []string) string {
	if s.Slack.Offline {
		return ""
	}

	result := ""
	for _, name := range names {
		result += s.fetchEmojiSingle(name)
	}

	return result
}

func (s *Server) fetchEmojiSingle(name string) string {
	if s.Slack.Offline {
		return ""
	}

	r, err := http.Get("https://e.benjaminsmith.dev/" + name)
	if err != nil {
		return ""
	}

	defer r.Body.Close()

	if r.StatusCode != 200 {
		return ""
	}

	return "![" + name + "](https://e.benjaminsmith.dev/" + name + ")"
}
