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

func makeDummyRoom() room {
	drw := dummyRW{}
	dummyTerm := terminal.NewTerminal(drw, "")
	ret := room{
		name:  "DummyRoom",
		users: []*user{},
	}
	tim := &user{
		name: "tim",
		term: dummyTerm,
	}
	tim.initColor()
	tim.changeColor("red")
	ret.users = append(ret.users, tim)
	tom := &user{
		name: "tom",
		term: dummyTerm,
	}
	tom.initColor()
	tom.changeColor("blue")
	ret.users = append(ret.users, tom)
	timtom := &user{
		name: "timtom",
		term: dummyTerm,
	}
	timtom.initColor()
	timtom.changeColor("sky")
	ret.users = append(ret.users, timtom)
	timt := &user{
		name: "timt",
		term: dummyTerm,
	}
	timt.initColor()
	timt.changeColor("coral")
	ret.users = append(ret.users, timt)
	return ret
}

func TestFindMention(t *testing.T) {
	r := makeDummyRoom()
	inputMsg := "@tim @tom @timtom @timt Hi!"
	// Warning, the order the elements have been put in the dummy room affects the result of the test
	expectedMsg := r.users[0].name + " " + r.users[1].name + " " + r.users[2].name + " " + r.users[3].name + " Hi!"
	coloredMsg := r.findMention(inputMsg)
	t.Log(coloredMsg)
	if coloredMsg != expectedMsg {
		t.Log(expectedMsg)
		t.Fail()
	}
	//fmt.Println(r.findMention("@tim @tom @timt @timtom Hi!"))
}
