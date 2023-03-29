package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/bwmarrin/discordgo"
	"github.com/leaanthony/go-ansi-parser"
	"github.com/quackduck/term"
	"golang.org/x/image/draw"
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
	sess.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentGuildWebhooks // listen to messages, manage webhooks
	err = sess.Open()
	if err != nil {
		Log.Println("Error opening Discord session:", err)
		return
	}

	var webhook *discordgo.Webhook
	// get or create a webhook if we're not in compact mode
	if !Integrations.Discord.CompactMode {
		webhooks, err := sess.ChannelWebhooks(Integrations.Discord.ChannelID)
		if err != nil {
			Log.Println("Error getting Discord webhooks:", err)
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
				Log.Println("Error creating a Discord webhook:", err)
				return
			}
		}
	}
	DiscordChan = make(chan DiscordMsg, 100)
	go func() {
		for msg := range DiscordChan {
			txt := strings.ReplaceAll(msg.msg, "@everyone", "@\\everyone")
			if Integrations.Discord.CompactMode {
				var toSend string
				if msg.senderName == "" {
					toSend = strings.ReplaceAll(stripansi.Strip("["+msg.channel+"] "+txt), `\n`, "\n")
				} else {
					toSend = strings.ReplaceAll(stripansi.Strip("["+msg.channel+"] **"+msg.senderName+"**: "+txt), `\n`, "\n")
				}
				_, err = sess.ChannelMessageSend(Integrations.Discord.ChannelID, toSend)
				if err != nil {
					Log.Println("Error sending Discord message:", err)
				}
			} else {
				_, err := sess.WebhookEditWithToken(webhook.ID, webhook.Token, webhook.Name, createDiscordImage(msg.senderName))
				if err != nil {
					Log.Println("Error modifying Discord webhook:", err)
				}
				_, err = sess.WebhookExecute(webhook.ID, webhook.Token, true,
					&discordgo.WebhookParams{
						Content:  strings.ReplaceAll(stripansi.Strip(txt), `\n`, "\n"),
						Username: stripansi.Strip("[" + msg.channel + "] " + msg.senderName),
					},
				)
				if err != nil {
					Log.Println("Error sending Discord message:", err)
				}
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

	msgContent := strings.TrimSpace(m.ContentWithMentionsReplaced())
	if Integrations.Slack != nil {
		SlackChan <- Integrations.Discord.Prefix + " " + m.Author.Username + ": " + msgContent // send this discord message to slack
	}
	runCommands(msgContent, DiscordUser)
}

var cacheSize = 20

// basic cache system
var imageCache = make([]struct {
	user  string
	image string
}, cacheSize)

func createDiscordImage(user string) string {
	// a completely transparent one pixel png
	fallback := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP//AA8AA/6P688AAAAASUVORK5CYII="
	if user == "" {
		// make messages with no sender (eg. command outputs) look seamless
		return fallback
	}
	for i := range imageCache {
		if imageCache[i].user == user {
			return imageCache[i].image
		}
	}
	styledTexts, err := ansi.Parse(user)
	if err != nil {
		Log.Println("Error parsing ANSI from username while creating Discord avatar:", err)
		return fallback
	}
	img := image.NewNRGBA(image.Rect(0, 0, len(styledTexts), 3))

	for i := 0; i < len(styledTexts); i++ {
		for j := 0; j < 3 && styledTexts[i].FgCol != nil; j++ {
			col := styledTexts[i].FgCol
			if (j == 0 || j == 2) && styledTexts[i].BgCol != nil {
				col = styledTexts[i].BgCol
			}
			img.Set(i, j, color.NRGBA{R: col.Rgb.R, G: col.Rgb.G, B: col.Rgb.B, A: 255})
		}
	}

	dst := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	draw.CatmullRom.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	//draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	buff := new(bytes.Buffer)
	err = png.Encode(buff, dst)
	if err != nil {
		Log.Println("Error creating Discord avatar:", err)
		return fallback
	}
	result := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buff.Bytes())

	if len(imageCache) >= cacheSize {
		// remove the first value
		imageCache = imageCache[1:]
	}
	imageCache = append(imageCache, struct {
		user  string
		image string
	}{user: user, image: result})
	return result
}
