package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/dghubble/go-twitter/twitter" //nolint:staticcheck // library deprecated
	"github.com/dghubble/oauth1"
)

var (
	StartupTime         = time.Now()
	Client              *twitter.Client
	AllowTweet          = true
	lastTwitterErrorDay time.Time // Track the day of the last Twitter error
)

func sendCurrentUsersTwitterMessage() {
	if Integrations.Twitter == nil {
		return
	}
	// TODO: count all users in all rooms
	if len(MainRoom.users) == 0 {
		return
	}
	if !AllowTweet {
		return
	}
	AllowTweet = false
	usersSnapshot := append(make([]*User, 0, len(MainRoom.users)), MainRoom.users...)
	areUsersEqual := func(a []*User, b []*User) bool {
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
		AllowTweet = true
		if !areUsersEqual(MainRoom.users, usersSnapshot) {
			return
		}
		Log.Println("Sending twitter update")
		names := make([]string, 0, len(MainRoom.users))
		for _, us := range MainRoom.users {
			names = append(names, us.Name)
		}
		tweetContent := "People on Devzat rn: " + stripansi.Strip(fmt.Sprint(names)) +
			"\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: " + printPrettyDuration(time.Since(StartupTime))

		// Try to post the tweet
		t, _, err := Client.Statuses.Update(tweetContent, nil)
		if err != nil {
			// Handle Twitter errors
			if isDuplicateStatusError(err) {
				Log.Println("Ignoring duplicate tweet.")
			} else if !isSameDay(time.Now(), lastTwitterErrorDay) {
				MainRoom.broadcast(Devbot, "Error sending tweet. Please try again later.")
				lastTwitterErrorDay = time.Now()
			}
			Log.Println("Failed to send tweet:", err)
			return
		}
		MainRoom.broadcast(Devbot, "https://twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr)
	}()
}

func twitterInit() { // called by init() in config.go
	if Integrations.Twitter == nil {
		return
	}

	config := oauth1.NewConfig(Integrations.Twitter.ConsumerKey, Integrations.Twitter.ConsumerSecret)
	token := oauth1.NewToken(Integrations.Twitter.AccessToken, Integrations.Twitter.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	Client = twitter.NewClient(httpClient)
	_, _, err := Client.Accounts.VerifyCredentials(nil)
	if err != nil {
		Log.Println("Twitter auth failed:", err)
		Integrations.Twitter = nil
	}
}
