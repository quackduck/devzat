// This module is meant to be some standard Go unit tests for Devzat. Run
// `go test` to run them.

package main

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	terminal "github.com/quackduck/term"
)

type dummyRW struct{}

func (rw dummyRW) Read(p []byte) (n int, err error)  { return 0, nil }
func (rw dummyRW) Write(p []byte) (n int, err error) { return 0, nil }

func makeDummyRoom() *Room {
	drw := dummyRW{}
	dummyTerm := terminal.NewTerminal(drw, "")
	ret := &Room{name: "DummyRoom", users: []*User{}}

	tim := &User{Name: "tim", term: dummyTerm, ColorBG: "bg-off"}
	_ = tim.changeColor("red")
	tom := &User{Name: "tom", term: dummyTerm, ColorBG: "bg-off"}
	_ = tom.changeColor("blue")
	timtom := &User{Name: "timtom", term: dummyTerm, ColorBG: "bg-off"}
	_ = timtom.changeColor("sky")
	timt := &User{Name: "timt", term: dummyTerm, ColorBG: "bg-off"}
	_ = timt.changeColor("coral")

	ret.users = append(ret.users, tim, tom, timtom, timt)
	return ret
}

/* ------------------ Testing correctness of findMention ------------------- */

func performTFM(t *testing.T, r *Room, raw string, expected string) {
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
	performTFM(t, r, inputMsg, expectedMsg)
	inputMsg = "@tim @fakemention"
	expectedMsg = r.users[0].Name + " @fakemention"
	performTFM(t, r, inputMsg, expectedMsg)
	inputMsg = "  @tim  "
	expectedMsg = "  " + r.users[0].Name + "  "
	performTFM(t, r, inputMsg, expectedMsg)
	performTFM(t, r, "", "")
	performTFM(t, r, "@tom", r.users[1].Name)
	performTFM(t, r, "no mention", "no mention")
	performTFM(t, r, "@tim \\@tim", r.users[0].Name+" @tim")
	performTFM(t, r, "@", "@")
	performTFM(t, r, "\\@", "@")
}

/* ---------------------- Testing speed of findMention ---------------------- */

var (
	noMention      = "This is a message with no mentions."
	compactMention = "@tom @tom @tom @tom."
	longMention    = "@tim, This is a l" + strings.Repeat("ooooo", 100) + "nger message"
	escapedMention = "\\@tom \\@tom \\@tom \\@tom."
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
