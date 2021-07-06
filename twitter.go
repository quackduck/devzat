package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/lunixbochs/vtclean"
)

var (
	client = loadTwitterClient()
)

// Credentials stores Twitter creds
type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func sendCurrentUsersTwitterMessage() {
	if offline {
		return
	}
	// TODO: count all users in all rooms
	if len(mainRoom.users) == 0 {
		return
	}
	usersSnapshot := append(make([]*user, 0, len(mainRoom.users)), mainRoom.users...)
	areUsersEqual := func(a []*user, b []*user) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if b[i] != a[i] {
				return false
			}
		}
		return true
	}
	go func() {
		time.Sleep(time.Second * 60)
		if !areUsersEqual(mainRoom.users, usersSnapshot) {
			return
		}
		l.Println("Sending twitter update")
		names := make([]string, 0, len(mainRoom.users))
		for _, us := range mainRoom.users {
			names = append(names, us.name)
		}
		t, _, err := client.Statuses.Update("People on Devzat rn: "+vtclean.Clean(fmt.Sprint(names), false)+"\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: "+printPrettyDuration(time.Since(startupTime)), nil)
		if err != nil {
			if !strings.Contains(err.Error(), "twitter: 187 Status is a duplicate.") {
				mainRoom.broadcast(devbot, "err: "+err.Error())
			}
			l.Println("Got twitter err", err)
			return
		}
		mainRoom.broadcast(devbot, "https\\://twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr)
	}()
}

func loadTwitterClient() *twitter.Client {
	if offline {
		return nil
	}
	d, err := ioutil.ReadFile("twitter-creds.json")
	if err != nil {
		panic(err)
	}
	twitterCreds := new(Credentials)
	err = json.Unmarshal(d, twitterCreds)
	if err != nil {
		panic(err)
	}
	config := oauth1.NewConfig(twitterCreds.ConsumerKey, twitterCreds.ConsumerSecret)
	token := oauth1.NewToken(twitterCreds.AccessToken, twitterCreds.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	t := twitter.NewClient(httpClient)
	return t
}
