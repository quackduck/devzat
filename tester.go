// This module is meant to perform some unit tests on Devzat code.

package main

import (
    "fmt"
	terminal "github.com/quackduck/term"
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
    ret := room {
        name:  "DummyRoom",
        users: []*user{},
    }
    tim := &user {
        name: "tim",
        term:  dummyTerm,
    }
    tim.initColor()
    tim.changeColor("red")
    ret.users = append(ret.users, tim);
    tom := &user {
        name: "tom",
        term:  dummyTerm,
    }
    tom.initColor()
    tom.changeColor("blue")
    ret.users = append(ret.users, tom);
    timtom := &user {
        name: "timtom",
        term:  dummyTerm,
    }
    timtom.initColor()
    timtom.changeColor("sky")
    timt := &user {
        name: "timt",
        term:  dummyTerm,
    }
    timt.initColor()
    timt.changeColor("coral")
    ret.users = append(ret.users, timt);
    ret.users = append(ret.users, timtom);
    return ret
}

func test() {
    r := makeDummyRoom()
    fmt.Println(r.findMention("@tim @tom @timt @timtom Hi!"))
}

