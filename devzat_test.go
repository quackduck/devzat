// This module is meant to be some standard Go unit tests for Devzat. Run
// `go test` to run them.

package main

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	terminal "github.com/quackduck/term"
)

const benchmarkRuns = 10000

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
	performTFM(t, r, "no mention", "no mention")
	performTFM(t, r, "@tim \\@tim", r.users[0].Name+" @tim")
	performTFM(t, r, "@", "@")
	performTFM(t, r, "\\@", "@")
}

/* ---------------------- Testing speed of findMention ---------------------- */

func oldMention(r *Room, msg string) string {
	for i := range r.users {
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(r.users[i].Name), r.users[i].Name)
		msg = strings.ReplaceAll(msg, `\`+r.users[i].Name, "@"+stripansi.Strip(r.users[i].Name)) // allow escaping
	}
	return msg
}

func BenchmarkFindMentionNoMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = r.findMention("This is a message with no mentions.")
	}
}

func BenchmarkFindMentionCompactMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = r.findMention("@tom @tom @tom @tom.")
	}
}

func BenchmarkFindMentionLongMessage(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = r.findMention("@tim, This is a looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooonger message")
	}
}

func BenchmarkFindMentionEscapedMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = r.findMention("\\@tom \\@tom \\@tom \\@tom.")
	}
}

func BenchmarkOldMentionNoMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = oldMention(r, "This is a message with no mentions.")
	}
}

func BenchmarkOldMentionCompactMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = oldMention(r, "@tom @tom @tom @tom.")
	}
}

func BenchmarkOldMentionLongMessage(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = oldMention(r, "@tim, This is a looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooonger message")
	}
}

func BenchmarkOldMentionEscapedMention(b *testing.B) {
	r := makeDummyRoom()
	for i := 0; i < benchmarkRuns; i++ {
		_ = oldMention(r, "\\@tom \\@tom \\@tom \\@tom.")
	}
}
