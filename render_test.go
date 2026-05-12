package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/quackduck/term"
)

func TestLineBreakIsCorrect(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	terminal := term.NewTerminal(buff, "")
	_ = terminal.SetSize(10000, 10000) // disable any formatting done by term
	user := &User{
		term:     terminal,
		winWidth: 80,
	}
	tests := []struct {
		name  string
		input Message
		want  string
	}{
		{
			name: "SimpleMessage",
			input: NewFakeUserMessage(
				"Test",
				"Testing",
				makeDummyRoom(),
			),
			want: "Test: Testing\r\n",
		},
		{
			name: "SimpleMessageLineWrap",
			input: NewFakeUserMessage(
				"Test",
				"Testing With A Long String That Is Over 80 chars long even excluding the username",
				makeDummyRoom(),
			),
			want: "Test: Testing With A Long String That Is Over 80 chars long even excluding the\r\n      username\r\n",
		},
		{
			name: "TestLineBreakIsCorrectOfByOne",
			input: NewFakeUserMessage(
				"Arkaeriit",
				"Maybe there is a off-by-one error in the way the line width is counted.",
				makeDummyRoom(),
			),
			want: "Arkaeriit: Maybe there is a off-by-one error in the way the line width is\r\n           counted.\r\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user.lastTimestamp = time.Now()
			user.writeln(test.input)
			read, err := io.ReadAll(buff)
			if err != nil {
				panic(err)
			}
			received := removeSpaceBeforeLineBreak(stripansi.Strip(string(read)))
			if received != test.want {
				t.Errorf("got %q, want %q", received, test.want)
			}
		})
	}
}

func removeSpaceBeforeLineBreak(s string) string {
	oldLines := strings.Split(s, "\r\n")
	lines := make([]string, 0, len(oldLines))
	for _, line := range oldLines {
		lines = append(lines, strings.TrimRight(line, " "))
	}

	return strings.Join(lines, "\r\n")
}
