package server

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	"devzat/pkg/room"
	_ "embed"
	"io"
	"log"
	_ "net/http/pprof"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/dghubble/go-twitter/twitter"
)

const (
	defaultLogFileName      = "log.txt"
	defaultAsciiArtFileName = "art.txt"
)

const (
	Scrollback       = 16
	NumStartingBans  = 10
	DefaultUserCount = 10
	HugeTerminalSize = 10000
)

const (
	tooManyLogins = 6
)

type Server struct {
	Admins AdminInfoMap

	Port        int
	Scrollback  int
	ProfilePort int

	// should this instance run offline? (should it not connect to slack or twitter?)
	OfflineSlack   bool
	OfflineTwitter bool

	mainRoom *room.Room
	Rooms    map[string]*room.Room

	Backlog          []models.BacklogMessage
	Bans             []models.Ban
	idsInMinToTimes  map[string]int
	AntiSpamMessages map[string]int

	Logfile io.WriteCloser
	Log     *log.Logger // prints to stdout as well as the logfile

	startupTime time.Time

	goaway.ProfanityDetector

	twitterClient *twitter.Client
	commands      map[string]interfaces.CommandFunc
}
