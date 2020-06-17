package main

import (
	"log"
	"os"
	"time"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	var (
		port      = os.Getenv("PORT")
		publicURL = os.Getenv("PUBLIC_URL") // you must add it to your config vars
		token     = os.Getenv("TOKEN")      // you must add it to your config vars
	)

	webhook := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	pref := tb.Settings{
		Token:  token,
		Poller: webhook,
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	b.Handle("/count", func(m *tb.Message) {
		b.Send(m.Chat, "3")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "2")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "1")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "Go!")
	})

	b.Start()
}
