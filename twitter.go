package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var (
	client     *twitter.Client
	allowTweet = true
)

func sendCurrentUsersTwitterMessage() {
	if Integrations.Twitter == nil {
		return
	}
	// TODO: count all users in all rooms
	if len(mainRoom.users) == 0 {
		return
	}
	if !allowTweet {
		return
	}
	allowTweet = false
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
		allowTweet = true
		if !areUsersEqual(mainRoom.users, usersSnapshot) {
			return
		}
		l.Println("Sending twitter update")
		names := make([]string, 0, len(mainRoom.users))
		for _, us := range mainRoom.users {
			names = append(names, us.Name)
		}
		t, _, err := client.Statuses.Update("People on Devzat rn: "+stripansi.Strip(fmt.Sprint(names))+"\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: "+printPrettyDuration(time.Since(startupTime)), nil)
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

func twitterInit() { // called by init() in config.go
	if Integrations.Twitter == nil {
		return
	}

	config := oauth1.NewConfig(Integrations.Twitter.ConsumerKey, Integrations.Twitter.ConsumerSecret)
	token := oauth1.NewToken(Integrations.Twitter.AccessToken, Integrations.Twitter.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client = twitter.NewClient(httpClient)
}
