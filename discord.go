package main

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/acarl005/stripansi"
	"github.com/bwmarrin/discordgo"
	"github.com/quackduck/term"
	"os"
	"strconv"
	"strings"
)

var (
	DiscordChan chan string
	DiscordUser = new(User)
)

func discordInit() {
	if Integrations.Discord == nil {
		return
	}

	sess, err := discordgo.New("Bot " + Integrations.Discord.Token)
	if err != nil {
		Log.Println("Error creating Discord session:", err)
		return
	}

	sess.AddHandler(discordMessageHandler)
	sess.Identify.Intents = discordgo.IntentsGuildMessages // only listen to messages
	err = sess.Open()
	if err != nil {
		Log.Println("Error opening Discord session:", err)
		return
	}

	DiscordChan = make(chan string, 100)
	go func() {
		for msg := range DiscordChan {
			msg = strings.ReplaceAll(msg, "@everyone", "@\\everyone")
			_, err = sess.ChannelMessageSend(Integrations.Discord.ChannelID, strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n"))
			if err != nil {
				Log.Println("Error sending Discord message:", err)
			}
		}
	}()

	DiscordUser.isBridge = true
	devnull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	DiscordUser.term = term.NewTerminal(devnull, "")
	DiscordUser.room = MainRoom
	Log.Println("Connected to Discord with bot ID", sess.State.User.ID, "as", sess.State.User.Username)
}

func discordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil || m.Author == nil || m.Author.ID == s.State.User.ID || m.ChannelID != Integrations.Discord.ChannelID { // ignore self and other channels
		return
	}
	h := sha1.Sum([]byte(m.Author.ID))
	i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
	DiscordUser.Name = Magenta.Paint(Integrations.Discord.Prefix+" ") + (Styles[int(i)%len(Styles)]).apply(m.Author.Username)

	msgContent := strings.TrimSpace(m.ContentWithMentionsReplaced())
	if Integrations.Slack != nil {
		SlackChan <- Integrations.Discord.Prefix + " " + m.Author.Username + ": " + msgContent // send this discord message to slack
	}
	runCommands(msgContent, DiscordUser)
}
