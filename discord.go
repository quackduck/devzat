package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/acarl005/stripansi"
)

var (
	discordChan = getSendToDiscordChan()
	webhook     = "https://discord.com/api/webhooks/<webhook_id>"
)

func getSendToDiscordChan() chan string {
	msgs := make(chan string, 100)
	go func() {
		for msg := range msgs {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			//if strings.HasPrefix(msg, "sshchat: ") { // just in case
			//	continue
			//}
			data, _ := json.Marshal(map[string]string{
				//"avatar_url": "",
				"username": "Discord Helper",
				"content":  msg,
			})
			http.Post(webhook, "application/json", bytes.NewBuffer(data))
		}
	}()
	return msgs
}
