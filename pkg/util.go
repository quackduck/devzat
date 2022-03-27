package pkg

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/acarl005/stripansi"
	markdown "github.com/quackduck/go-term-markdown"
)

const (
	defaultAdminsFileName   = "admins.json"
	defaultAsciiArtFileName = "art.txt"
)

const (
	fmtAppendNewline = "%s\n%s"
)

func getASCIIArt() string {
	b, _ := ioutil.ReadFile("art.txt")
	if b == nil {
		return "sowwy, no art was found, please slap your developer and tell em to add an art.txt file"
	}
	return string(b)
}

func formatNames(names []string) string {
	joined := strings.Join(names, " ")

	return fmt.Sprintf("[%s]", joined)
}

func lenString(a string) int {
	return len([]rune(stripansi.Strip(a)))
}

func autogenCommands(cmds []Command) string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)

	for _, c := range cmds {
		formatted := fmt.Sprintf("   %s\t%s\t_%s_  \n", c.name, c.argsInfo, c.info)
		_, _ = w.Write([]byte(formatted))
	}

	_ = w.Flush()

	return b.String()
}

// removes arrows, spaces and non-ascii-printable characters
func cleanName(name string) string {
	s := ""
	name = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		strings.TrimSpace(strings.Split(name, "\n")[0]), // use one trimmed line
		"<-", ""),
		"->", ""),
		" ", "-")
	if len([]rune(name)) > 27 {
		name = string([]rune(name)[:27])
	}
	for i := 0; i < len(name); i++ {
		if 33 <= name[i] && name[i] <= 126 { // ascii printables only: '!' to '~'
			s += string(name[i])
		}
	}
	return s
}

func printPrettyDuration(d time.Duration) string {
	s := strings.TrimSpace(strings.TrimSuffix(d.Round(time.Minute).String(), "0s"))
	if s == "" { // we cut off the seconds so if there's nothing in the string it means it was made of only seconds.
		s = "< 1m"
	}

	return s
}

func mdRender(a string, beforeMessageLen int, lineWidth int) string {
	if strings.Contains(a, "![") && strings.Contains(a, "](") {
		lineWidth = int(math.Min(float64(lineWidth/2), 200)) // max image width is 200
	}

	md := string(markdown.Render(a, lineWidth-(beforeMessageLen), 0))
	md = strings.TrimSuffix(md, "\n")
	split := strings.Split(md, "\n")

	for i := range split {
		if i == 0 {
			continue // the first line will automatically be padded
		}

		split[i] = strings.Repeat(" ", beforeMessageLen) + split[i]
	}

	if len(split) == 1 {
		return md
	}

	return strings.Join(split, "\n")
}

// Returns true and the User with the same name if the username is taken, false and nil otherwise

func remove(s []*User, a *User) []*User {
	for j := range s {
		if s[j] == a {
			return append(s[:j], s[j+1:]...)
		}
	}
	return s
}

func shasum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
