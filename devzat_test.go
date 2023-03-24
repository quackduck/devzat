// This module is meant to be some standard Go unit tests for Devzat. Run
// `go test` to run them.

package main

import (
	"io"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"
)

type dummyRW struct{}

func (rw dummyRW) Read(p []byte) (n int, err error)  { return 0, nil }
func (rw dummyRW) Write(p []byte) (n int, err error) { return 0, nil }

type dummySession struct{}

func (s dummySession) User() string                            { return "" }
func (s dummySession) RemoteAddr() net.Addr                    { return nil }
func (s dummySession) LocalAddr() net.Addr                     { return nil }
func (s dummySession) Environ() []string                       { return make([]string, 0) }
func (s dummySession) Exit(code int) error                     { return nil }
func (s dummySession) Command() []string                       { return make([]string, 0) }
func (s dummySession) RawCommand() string                      { return "" }
func (s dummySession) Subsystem() string                       { return "" }
func (s dummySession) PublicKey() ssh.PublicKey                { return nil }
func (s dummySession) Context() ssh.Context                    { return nil }
func (s dummySession) Permissions() ssh.Permissions            { return ssh.Permissions{} }
func (s dummySession) Pty() (ssh.Pty, <-chan ssh.Window, bool) { return ssh.Pty{}, nil, true }
func (s dummySession) Signals(c chan<- ssh.Signal)             {}
func (s dummySession) Break(c chan<- bool)                     {}
func (s dummySession) Close() error                            { return nil }
func (s dummySession) CloseWrite() error                       { return nil }
func (s dummySession) Read(data []byte) (int, error)           { return 0, nil }
func (s dummySession) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return false, nil
}
func (s dummySession) Stderr() io.ReadWriter          { return nil }
func (s dummySession) Write(data []byte) (int, error) { return 0, nil }

func makeDummyRoom() *Room {
	drw := dummyRW{}
	dummyTerm := terminal.NewTerminal(drw, "")
	ret := &Room{name: "DummyRoom", users: []*User{}, usersMutex: sync.RWMutex{}}

	tim := &User{Name: "tim", term: dummyTerm, ColorBG: "bg-off", room: ret, session: dummySession{}}
	_ = tim.changeColor("red")
	tom := &User{Name: "tom", term: dummyTerm, ColorBG: "bg-off", room: ret, session: dummySession{}}
	_ = tom.changeColor("blue")
	timtom := &User{Name: "timtom", term: dummyTerm, ColorBG: "bg-off", room: ret, session: dummySession{}}
	_ = timtom.changeColor("sky")
	timt := &User{Name: "timt", term: dummyTerm, ColorBG: "bg-off", room: ret, session: dummySession{}}
	_ = timt.changeColor("coral")

	ret.users = append(ret.users, tim, tom, timtom, timt)
	return ret
}

/* ------------------ Testing correctness of findMention ------------------- */

func performTestFindMention(t *testing.T, r *Room, raw string, expected string) {
	colored := r.findMention(raw)
	t.Log(colored)
	if colored != expected {
		t.Log(expected)
		t.Fail()
	}
}

func TestFindMention(t *testing.T) {
	r := makeDummyRoom()
	inputMsg := "@tim @tom @timtom @timt Hi!"
	// Warning, the order the elements have been put in the dummy room affects the result of the test
	expectedMsg := r.users[0].Name + " " + r.users[1].Name + " " + r.users[2].Name + " " + r.users[3].Name + " Hi!"
	performTestFindMention(t, r, inputMsg, expectedMsg)
	inputMsg = "@tim @fakemention"
	expectedMsg = r.users[0].Name + " @fakemention"
	performTestFindMention(t, r, inputMsg, expectedMsg)
	inputMsg = "  @tim  "
	expectedMsg = "  " + r.users[0].Name + "  "
	performTestFindMention(t, r, inputMsg, expectedMsg)
	performTestFindMention(t, r, "", "")
	performTestFindMention(t, r, "@tom", r.users[1].Name)
	performTestFindMention(t, r, "no mention", "no mention")
	performTestFindMention(t, r, "@tim \\@tim", r.users[0].Name+" @tim")
	performTestFindMention(t, r, "@", "@")
	performTestFindMention(t, r, "\\@", "@")
}

/* ---------------------- Testing speed of findMention ---------------------- */

var (
	noMention      = strings.Repeat("This is a message with no mentions.", 100)
	compactMention = "@timt @timt @tom @timt."
	longMention    = strings.Repeat(" @timt", 100)
	escapedMention = strings.Repeat(" \\@timt", 100)
)

func oldMention(r *Room, msg string) string {
	for i := range r.users {
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(r.users[i].Name), r.users[i].Name)
		msg = strings.ReplaceAll(msg, `\`+r.users[i].Name, "@"+stripansi.Strip(r.users[i].Name)) // allow escaping
	}
	return msg
}

func BenchmarkFindMentionNoMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = r.findMention(noMention)
	}
}

func BenchmarkFindMentionCompactMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = r.findMention(compactMention)
	}
}

func BenchmarkFindMentionLongMessage(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = r.findMention(longMention)
	}
}

func BenchmarkFindMentionEscapedMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = r.findMention(escapedMention)
	}
}

func BenchmarkOldMentionNoMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = oldMention(r, noMention)
	}
}

func BenchmarkOldMentionCompactMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = oldMention(r, compactMention)
	}
}

func BenchmarkOldMentionLongMessage(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = oldMention(r, longMention)
	}
}

func BenchmarkOldMentionEscapedMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < b.N; i++ {
		_ = oldMention(r, escapedMention)
	}
}

/* --------------------- Testing correctness of banning --------------------- */

func performTestBan(t *testing.T, id0 string, id1 string, id2 string, id3 string, usersBanned int) {
	r := makeDummyRoom()
	Rooms = map[string]*Room{r.name: r}
	r.users[0].id = id0
	r.users[1].id = id1
	r.users[2].id = id2
	r.users[3].id = id3
	r.users[0].ban("Tim is a meany")
	if len(r.users) != 4-usersBanned {
		t.Log("Error,", usersBanned, "users should have been kicked but", 4-len(r.users), "have been kicked.")
		t.Fail()
	}
}

func TestBan(t *testing.T) {
	// Testing a single user being banned
	performTestBan(t, "0", "1", "2", "3", 1)
	// Testing the two consecutive users sharing the same ID
	performTestBan(t, "bad", "bad", "900d", "900d", 2)
	// Testing all user having the same id
	performTestBan(t, "same", "same", "same", "same", 4)
	// Testing interlaced users sharing the same ID
	performTestBan(t, "bad", "900d", "bad", "900d", 2)
}
