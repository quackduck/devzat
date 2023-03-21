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
	DiscordChan chan DiscordMsg
	DiscordUser = new(User)
)

type DiscordMsg struct {
	senderName string
	msg        string
	channel    string
}

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

	var webhook *discordgo.Webhook
	//get or create Webhook
	if Integrations.Discord.DiscordStyleUsername {
		webhooks, err := sess.ChannelWebhooks(Integrations.Discord.ChannelID)
		if err != nil {
			Log.Println("Error getting Webhooks:", err)
			return
		}
		for _, wh := range webhooks {
			if wh.Name == "Devzat" {
				webhook = wh
			}
		}
		if webhook == nil {
			webhook, err = sess.WebhookCreate(Integrations.Discord.ChannelID, "Devzat", "")
			if err != nil {
				Log.Println("Error creating Webhook:", err)
				return
			}
		}
	}
	DiscordChan = make(chan DiscordMsg, 100)
	go func() {
		for msg := range DiscordChan {
			txt := strings.ReplaceAll(msg.msg, "@everyone", "@\\everyone")
			if Integrations.Discord.DiscordStyleUsername {
				_, err = sess.WebhookExecute(
					webhook.ID,
					webhook.Token,
					true,
					&discordgo.WebhookParams{ //TODO: maybe change the pfp based on the Sender (color?)
						Content:  strings.ReplaceAll(stripansi.Strip(txt), `\n`, "\n"),
						Username: stripansi.Strip("[" + msg.channel + "] " + msg.senderName),
					},
				)
			} else {
				var toSend string
				if msg.senderName == "" {
					toSend = strings.ReplaceAll(stripansi.Strip("["+msg.channel+"] "+txt), `\n`, "\n")
				} else {
					toSend = strings.ReplaceAll(stripansi.Strip("["+msg.channel+"] **"+msg.senderName+"**: "+txt), `\n`, "\n")
				}
				_, err = sess.ChannelMessageSend(Integrations.Discord.ChannelID, toSend)
			}
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

func discordMessageHandler(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil || m.Author == nil || m.Author.Bot || m.ChannelID != Integrations.Discord.ChannelID { // ignore self and other channels
		return
	}
	h := sha1.Sum([]byte(m.Author.ID))
	i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
	DiscordUser.Name = Magenta.Paint(Integrations.Discord.Prefix+" ") + (Styles[int(i)%len(Styles)]).apply(m.Author.Username)
	m.Content = strings.TrimSpace(m.Content) // mildly cursed but eh who cares
	if Integrations.Slack != nil {
		SlackChan <- Integrations.Discord.Prefix + " " + m.Author.Username + ": " + m.Content // send this discord message to slack
	}
	runCommands(m.Content, DiscordUser)
}
