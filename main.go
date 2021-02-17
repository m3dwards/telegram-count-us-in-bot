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

	var (
		// Universal markup builders.
		menu     = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
		selector = &tb.ReplyMarkup{}

		// Reply buttons.
		btnHelp     = menu.Text("ℹ Help")
		btnSettings = menu.Text("⚙ Settings")

		// Inline buttons.
		//
		// Pressing it will cause the client to
		// send the bot a callback.
		//
		// Make sure Unique stays unique as per button kind,
		// as it has to be for callback routing to work.
		//
		btnPrev = selector.Data("⬅", "prev", "1")
		btnNext = selector.Data("➡", "next", "1")
	)

	menu.Reply(
		menu.Row(btnHelp),
		menu.Row(btnSettings),
	)
	selector.Inline(
		selector.Row(btnPrev, btnNext),
	)

	// Command: /start <PAYLOAD>
	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Chat, "Hello!", selector)

		// if !m.Private() {
		// 	return
		// }

		// b.Send(m.Sender, "Hello!", menu)
	})

	// On reply button pressed (message)
	b.Handle(&btnHelp, func(m *tb.Message) {})

	// On inline button pressed (callback)
	b.Handle(&btnPrev, func(c *tb.Callback) {
		// ...
		// Always respond!
		b.Respond(c, &tb.CallbackResponse{Text: "Previous"})
	})

	b.Handle(&btnNext, func(c *tb.Callback) {
		// ...
		// Always respond!
		b.Respond(c, &tb.CallbackResponse{Text: "Next"})
	})

	b.Handle("/count", func(m *tb.Message) {
		b.Send(m.Chat, "3")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "2")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "1")
		time.Sleep(1 * time.Second)
		b.Send(m.Chat, "Go!")
	})

	b.Handle("/playstation", func(m *tb.Message) {
		b.Send(m.Chat, "P")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "L")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "A")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "Y")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "S")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "T")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "A")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "T")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "I")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "O")
		time.Sleep(150 * time.Millisecond)
		b.Send(m.Chat, "N")
		time.Sleep(150 * time.Millisecond)
	})

	b.Handle("/llama", func(m *tb.Message) {
		a := &tb.Photo{File: tb.FromURL("https://pbs.twimg.com/profile_images/378800000802823295/fa4f4104d718899ea49f3a507c7f6034_400x400.jpeg")}
		if err != nil {
			return
		}
		b.Send(m.Chat, a)
	})

	b.Handle("/randomllama", func(m *tb.Message) {
		a := &tb.Photo{File: tb.FromURL("https://source.unsplash.com/800x600?llama")}
		if err != nil {
			return
		}
		b.Send(m.Chat, a)
	})

	b.Start()
}
