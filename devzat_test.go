// This module is meant to perform some unit tests on Devzat code.

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

func makeDummyRoom() Room {
	drw := dummyRW{}
	dummyTerm := terminal.NewTerminal(drw, "")
	ret := Room{
		name:  "DummyRoom",
		users: []*User{},
	}
	tim := &User{
		Name: "tim",
		term: dummyTerm,
	}
	tim.changeColor("red")
	ret.users = append(ret.users, tim)
	tom := &User{
		Name: "tom",
		term: dummyTerm,
	}
	tom.changeColor("blue")
	ret.users = append(ret.users, tom)
	timtom := &User{
		Name: "timtom",
		term: dummyTerm,
	}
	timtom.changeColor("sky")
	ret.users = append(ret.users, timtom)
	timt := &User{
		Name: "timt",
		term: dummyTerm,
	}
	timt.changeColor("coral")
	ret.users = append(ret.users, timt)
	return ret
}

func TestFindMention(t *testing.T) {
	r := makeDummyRoom()
	inputMsg := "@tim @tom @timtom @timt Hi!"
	// Warning, the order the elements have been put in the dummy room affects the result of the test
	expectedMsg := r.users[0].Name + " " + r.users[1].Name + " " + r.users[2].Name + " " + r.users[3].Name + " Hi!"
	coloredMsg := r.findMention(inputMsg)
	t.Log(coloredMsg)
	if coloredMsg != expectedMsg {
		t.Log(expectedMsg)
		t.Fail()
	}
	//fmt.Println(r.findMention("@tim @tom @timt @timtom Hi!"))
}
