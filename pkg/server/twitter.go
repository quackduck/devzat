package server

import (
	"devzat/pkg"
	"devzat/pkg/user"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var (
	allowTweet = true
)

type TwitterCreds struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func (s *Server) SendCurrentUsersTwitterMessage() {
	if s.OfflineTwitter {
		return
	}

	// TODO: count all users in all Rooms
	if len(s.MainRoom.Users) == 0 {
		return
	}

	if !allowTweet {
		return
	}
	allowTweet = false
	usersSnapshot := append(make([]*user.User, 0, len(s.MainRoom.Users)), s.MainRoom.Users...)
	areUsersEqual := func(a []*user.User, b []*user.User) bool {
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
		if !areUsersEqual(s.MainRoom.Users, usersSnapshot) {
			return
		}
		s.Log.Println("Sending twitter update")
		names := make([]string, 0, len(s.MainRoom.Users))
		for _, us := range s.MainRoom.Users {
			names = append(names, us.Name)
		}
		t, _, err := s.twitterClient.Statuses.Update("People on Devzat rn: "+stripansi.Strip(fmt.Sprint(names))+"\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: "+pkg.PrintPrettyDuration(time.Since(s.startupTime)), nil)
		if err != nil {
			if !strings.Contains(err.Error(), "twitter: 187 Status is a duplicate.") {
				s.MainRoom.Broadcast(s.MainRoom.Bot.Name(), "err: "+err.Error())
			}
			s.Log.Println("Got twitter err", err)
			return
		}
		s.MainRoom.Broadcast(s.MainRoom.Bot.Name(), "https\\://twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr)
	}()
}

func (s *Server) loadTwitterClient() {
	d, err := ioutil.ReadFile("twitter-creds.json")

	if os.IsNotExist(err) {
		s.OfflineTwitter = true
		s.Log.Println("Did not find twitter-creds.json. Enabling offline mode.")
	} else if err != nil {
		panic(err)
	}

	if s.OfflineTwitter {
		return
	}

	twitterCreds := new(TwitterCreds)
	err = json.Unmarshal(d, twitterCreds)
	if err != nil {
		panic(err)
	}
	config := oauth1.NewConfig(twitterCreds.ConsumerKey, twitterCreds.ConsumerSecret)
	token := oauth1.NewToken(twitterCreds.AccessToken, twitterCreds.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	s.twitterClient = twitter.NewClient(httpClient)
}
