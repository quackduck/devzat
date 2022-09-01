// This module is meant to perform some unit tests on Devzat code.

package main

import "fmt"

func makeDummyRoom() room {
    ret := room {
        name:  "DummyRoom",
        users: []*user{},
    }
    tim := &user {
        name: "tim",
    }
    tim.initColor()
    tim.changeColor("red")
    ret.users = append(ret.users, tim);
    tom := &user {
        name: "tom",
    }
    tom.initColor()
    tom.changeColor("blue")
    ret.users = append(ret.users, tom);
    timtom := &user {
        name: "timtom",
    }
    timtom.initColor()
    timtom.changeColor("sky")
    timt := &user {
        name: "timt",
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

