package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"github.com/acarl005/stripansi"
	"github.com/bwmarrin/discordgo"
	"github.com/leaanthony/go-ansi-parser"
	"github.com/quackduck/term"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"image/jpeg"
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
				avatar := createDiscordImage(msg.senderName)
				_, err := sess.WebhookEditWithToken(webhook.ID, webhook.Token, webhook.Name, avatar)
				if err != nil {
					Log.Println("Error modifying Discord webhook:", err)
				}
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

func createDiscordImage(user string) string {
	if user == "" {
		return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAAAXNSR0IArs4c6QAAAAxJREFUCJljMDayAAABOQCeivIjywAAAABJRU5ErkJggg=="
	}
	styledTexts, err := ansi.Parse(user)
	if err != nil {
		Log.Println("Error parsing username while creating Image.")
		return ""
	}
	_ = styledTexts
	img := image.NewRGBA(image.Rectangle{
		Min: image.Point{},
		Max: image.Point{X: len(styledTexts), Y: 5},
	})
	i := 0
	for i < len(styledTexts) {
		j := 0
		for j < 5 {
			if (j == 0 || j == 4) && styledTexts[i].BgCol != nil {
				img.Set(i, j, color.RGBA{
					R: styledTexts[i].BgCol.Rgb.R,
					G: styledTexts[i].BgCol.Rgb.G,
					B: styledTexts[i].BgCol.Rgb.B,
				})
			} else {
				img.Set(i, j, color.RGBA{
					R: styledTexts[i].FgCol.Rgb.R,
					G: styledTexts[i].FgCol.Rgb.G,
					B: styledTexts[i].FgCol.Rgb.B,
				})
			}
			j++
		}
		i++
	}

	dst := image.NewRGBA(image.Rect(0, 0, 128, 128))
	draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	var buff bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buff)
	err = jpeg.Encode(encoder, dst, nil)
	if err != nil {
		Log.Println("Error creating Image.")
		return ""
	}
	return "data:image/jpeg;base64," + buff.String()
}
