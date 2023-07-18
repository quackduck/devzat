package main

import (
	"fmt"
	"strings"
	"time"
	"os"

	"github.com/acarl005/stripansi"
	"github.com/dghubble/go-twitter/twitter" //nolint:staticcheck // library deprecated
	"github.com/dghubble/oauth1"
)

var (
	StartupTime = time.Now()
	Client      *twitter.Client
	AllowTweet  = true
)

const twitterLogFilename = "twitter_logs.txt"

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
		tweetText := "People on Devzat rn: " + stripansi.Strip(fmt.Sprint(names)) + "\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: " + printPrettyDuration(time.Since(StartupTime))

		t, _, err := Client.Statuses.Update(tweetText, nil)
		if err != nil {
			if !strings.Contains(err.Error(), "twitter: 187 Status is a duplicate.") {
				// Log the error to the file
				logErrorToFile(err)
			}
			Log.Println("Got twitter err", err)
			return
		}
		MainRoom.broadcast(Devbot, "https\\://twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr)
	}()
}

func logErrorToFile(err error) {
	file, err := os.OpenFile(twitterLogFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Log.Println("Error opening twitter log file:", err)
		return
	}
	defer file.Close()

	logTime := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] Error: %s\n", logTime, err.Error())

	if _, err = file.WriteString(logEntry); err != nil {
		Log.Println("Error writing to twitter log file:", err)
	}
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
