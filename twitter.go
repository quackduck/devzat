package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
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
	// TODO: count all users in all rooms
	if len(mainRoom.users) == 0 {
		return
	}
	usersSnapshot := mainRoom.users
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
		time.Sleep(time.Second * 30)
		if !areUsersEqual(mainRoom.users, usersSnapshot) {
			return
		}
		l.Println("Sending twitter update")
		//broadcast(devbot, "sending twitter update", true)
		names := make([]string, 0, len(mainRoom.users))
		for _, us := range mainRoom.users {
			names = append(names, us.name)
		}
		t, _, err := client.Statuses.Update("People on Devzat rn: "+stripansi.Strip(fmt.Sprint(names))+"\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: "+printPrettyDuration(time.Since(startupTime)), nil)
		if err != nil && err.Error() != "twitter: 187 Status is a duplicate." {
			l.Println("Got twitter err", err)
			mainRoom.broadcast(devbot, "err: "+err.Error(), true)
			return
		}
		mainRoom.broadcast(devbot, "https\\://twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr, true)
	}()
	//broadcast(devbot, tweet.Entities.Urls)
}

func loadTwitterClient() *twitter.Client {
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
