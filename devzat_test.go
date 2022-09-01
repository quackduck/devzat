// This module is meant to be some standard Go unit tests for Devzat. Run
// `go test` to run them.

package main

import (
	terminal "github.com/quackduck/term"
	"testing"
)

type dummyRW struct {
}

func (rw dummyRW) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (rw dummyRW) Write(p []byte) (n int, err error) {
	return 0, nil
}

func makeDummyRoom() *Room {
	drw := dummyRW{}
	dummyTerm := terminal.NewTerminal(drw, "")
	ret := &Room{
		name:  "DummyRoom",
		users: []*User{},
	}
	tim := &User{
		Name:    "tim",
		term:    dummyTerm,
		ColorBG: "bg-off",
	}
	_ = tim.changeColor("red")
	ret.users = append(ret.users, tim)
	tom := &User{
		Name:    "tom",
		term:    dummyTerm,
		ColorBG: "bg-off",
	}
	_ = tom.changeColor("blue")
	ret.users = append(ret.users, tom)
	timtom := &User{
		Name:    "timtom",
		term:    dummyTerm,
		ColorBG: "bg-off",
	}
	_ = timtom.changeColor("sky")
	ret.users = append(ret.users, timtom)
	timt := &User{
		Name:    "timt",
		term:    dummyTerm,
		ColorBG: "bg-off",
	}
	_ = timt.changeColor("coral")
	ret.users = append(ret.users, timt)
	return ret
}

func performTFM(t *testing.T, r *Room, raw string, expected string) {
	colored := r.findMention(raw)
	t.Log(colored)
	if colored != expected {
		t.Log(expected)
		t.Fail()
	}
}

// Test the findMention function
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
}
