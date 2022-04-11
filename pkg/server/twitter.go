package server

import (
	i "devzat/pkg/interfaces"
	"encoding/json"
	"fmt"
	"github.com/dghubble/oauth1"
	"io/ioutil"
	"strings"
	"time"

	"devzat/pkg/util"
	"github.com/acarl005/stripansi"
	"github.com/dghubble/go-twitter/twitter"
)

const (
	defaultTwitterConfigFile = "twitter-creds.json"
)

type twitterIntegration struct {
	allowTweet bool
	client     *twitter.Client
	creds      struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}
}

func (t *twitterIntegration) init() error {
	t.allowTweet = true

	d, err := ioutil.ReadFile(defaultTwitterConfigFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, &t.creds)
	if err != nil {
		return err
	}

	config := oauth1.NewConfig(t.creds.ConsumerKey, t.creds.ConsumerSecret)
	token := oauth1.NewToken(t.creds.AccessToken, t.creds.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	t.client = twitter.NewClient(httpClient)

	return nil
}

func (s *Server) SendCurrentUsersTwitterMessage() {
	return
	usersInMain := s.mainRoom.AllUsers()
	numUsersInMain := len(usersInMain)

	if numUsersInMain == 0 {
		return
	}

	if !s.twitter.allowTweet {
		return
	}

	s.twitter.allowTweet = false
	usersSnapshot := append(make([]i.User, 0, numUsersInMain), usersInMain...)
	areUsersEqual := func(a []i.User, b []i.User) bool {
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
		s.twitter.allowTweet = true

		if areUsersEqual(s.mainRoom.AllUsers(), usersSnapshot) {
			return
		}

		s.Info().Msg("Sending twitter update")
		names := make([]string, 0, len(s.MainRoom().AllUsers()))
		for _, us := range s.MainRoom().AllUsers() {
			names = append(names, us.Name())
		}

		t, _, err := s.twitter.client.Statuses.Update("People on Devzat rn: "+stripansi.Strip(fmt.Sprint(names))+"\nJoin em with \"ssh devzat.hackclub.com\"\nUptime: "+util.PrintPrettyDuration(time.Since(s.startupTime)), nil)
		if err != nil {
			if !strings.Contains(err.Error(), "twitter: 187 Status is a duplicate.") {
				s.mainRoom.Broadcast(s.mainRoom.Bot().Name(), "err: "+err.Error())
			}
			s.Error().Msgf("Got twitter err: %v", err)
			return
		}

		s.mainRoom.Broadcast(s.mainRoom.Bot().Name(), "https\\://twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr)
	}()
}
