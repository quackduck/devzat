package v2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"devzat/pkg/util"
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
	if s.settings.Twitter.Offline {
		return
	}

	// TODO: count all users in all Rooms
	if len(s.mainRoom.Users) == 0 {
		return
	}

	if !allowTweet {
		return
	}
	allowTweet = false
	usersSnapshot := append(make([]*user.User, 0, len(s.mainRoom.Users)), s.mainRoom.Users...)
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
		if !areUsersEqual(s.mainRoom.Users, usersSnapshot) {
			return
		}
		s.Log.Println("Sending twitter update")
		names := make([]string, 0, len(s.mainRoom.Users))
		for _, us := range s.mainRoom.Users {
			names = append(names, us.Name)
		}
		t, _, err := s.twitterClient.Statuses.Update("People on Devzat rn: "+stripansi.Strip(fmt.Sprint(names))+"\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: "+util.PrintPrettyDuration(time.Since(s.startupTime)), nil)
		if err != nil {
			if !strings.Contains(err.Error(), "twitter: 187 Status is a duplicate.") {
				s.mainRoom.Broadcast(s.mainRoom.Bot.Name(), "err: "+err.Error())
			}
			s.Log.Println("Got twitter err", err)
			return
		}
		s.mainRoom.Broadcast(s.mainRoom.Bot.Name(), "https\\://twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr)
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
