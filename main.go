package main

import (
	"log"
	"os"
	"strings"
	"time"
	// "strconv"
	"github.com/pborman/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

type viewer struct {
	Name     string
	Username string
	ID       int
	Ready    bool
	Timer *time.Timer
}

type watchParty struct {
	ID      string
	Name    string
	Viewers []*viewer
	OwnerID int
}

type replyID struct {
	ChatID int64
	MsgID  int
}

var data []*watchParty
var replyIDs []*replyID

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

	replyquery := &tb.ReplyMarkup{ForceReply: true, Selective: true}

	b.Handle("/start", func(m *tb.Message) {
		filmName := m.Text[6:]
		filmName = strings.TrimSpace(filmName)
		if len(filmName) == 0 {
			rep, _ := b.Send(m.Chat, "@"+m.Sender.Username+" enter the film or show name:", replyquery)
			addNewReplyId(m.Chat.ID, rep.ID)
			return
		}
		handleNewWatchParty(b, filmName, m.Sender.ID, m.Chat)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		if m.ReplyTo != nil && checkReplyIDExists(m.Chat.ID, m.ReplyTo.ID) {
			deleteReplyID(m.Chat.ID, m.ReplyTo.ID)
			filmName := strings.TrimSpace(m.Text)
			if len(filmName) == 0 {
				return
			}
			handleNewWatchParty(b, filmName, m.Sender.ID, m.Chat)
		}
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

func addNewReplyId(chatID int64, msgID int) {
	replyIDs = append(replyIDs, &replyID{ChatID: chatID, MsgID: msgID})
}

func checkReplyIDExists(chatID int64, msgID int) bool {
	for _, r := range replyIDs {
		if chatID == r.ChatID && msgID == r.MsgID {
			return true
		}
	}
	return false
}

func deleteReplyID(chatID int64, msgID int) {
	for i, r := range replyIDs {
		if chatID == r.ChatID && msgID == r.MsgID {
			replyIDs = append(replyIDs[:i], replyIDs[i+1:]...)
		}
	}
}

func getWatchPartyByID(ID string) *watchParty {
	for _, wp := range data {
		if ID == wp.ID {
			return wp
		}
	}
	return nil
}

func createNewWatchParty(name string, ownerID int) string {
	id := uuid.New()
	wp := &watchParty{ID: id, Name: name, OwnerID: ownerID}
	data = append(data, wp)
	return id
}

func handleNewWatchParty(b *tb.Bot, filmName string, senderID int, chat *tb.Chat) {
	b.Send(chat, "Who would like to watch "+filmName+"?")
	wpID := createNewWatchParty(filmName, senderID)

	InOrOut := &tb.ReplyMarkup{}
	btnIn := InOrOut.Data("I'm in!", "in", wpID)
	btnOut := InOrOut.Data("I'm not in", "out", wpID)
	btnInitiate := InOrOut.Data("Start countdown!", "initiate", wpID)
	InOrOut.Inline(InOrOut.Row(btnIn, btnOut), InOrOut.Row(btnInitiate))

	m, _ := b.Send(chat, "Nobody is in", InOrOut)

	b.Handle(&btnIn, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are in!"})
		wp := getWatchPartyByID(c.Data)
		addPersonToWP(wp, c.Sender.FirstName, c.Sender.Username, c.Sender.ID)
		b.Edit(m, getInOutMsg(wp), InOrOut)
	})
	b.Handle(&btnOut, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Removing you from watch party"})
		wp := getWatchPartyByID(c.Data)
		removeViewerFromWP(wp, &viewer{ID: c.Sender.ID})
		b.Edit(m, getInOutMsg(wp), InOrOut)
	})

	b.Handle(&btnInitiate, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Initiating countdown"})

		readyNotReady := &tb.ReplyMarkup{}
		btnReady := readyNotReady.Data("Paused and Ready!", "ready", wpID)
		btnNotReady := readyNotReady.Data("Not ready!", "notready", wpID)
		readyNotReady.Inline(InOrOut.Row(btnReady, btnNotReady))

		mr, _ := b.Send(chat, getReadyMsg(wpID), readyNotReady)

		b.Handle(&btnReady, func(c *tb.Callback) {
			b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are ready!"})
			setViewerStatus(c.Data, c.Sender.ID, true, mr, b, readyNotReady)
			if checkIfWeAreAGo(c.Data) {
				b.Send(m.Chat, "Looks like we are all ready! Starting timer.")
				time.Sleep(2 * time.Second)
				b.Send(m.Chat, "3")
				time.Sleep(1 * time.Second)
				b.Send(m.Chat, "2")
				time.Sleep(1 * time.Second)
				b.Send(m.Chat, "1")
				time.Sleep(1 * time.Second)
				b.Send(m.Chat, "Go!")
			}
		})
		b.Handle(&btnNotReady, func(c *tb.Callback) {
			b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are not ready"})
			setViewerStatus(c.Data, c.Sender.ID, false, mr, b, readyNotReady)
		})
	})
}

func checkIfWeAreAGo(wpID string) bool {
	wp := getWatchPartyByID(wpID)
	for _, v := range wp.Viewers {
		if !v.Ready {
			return false
		}
	}
	return true
}

func getViewerName(v *viewer) string {
	if len(v.Name) > 0 {
		return v.Name
	}
	return "@" + v.Username
}

func getInOutMsg(wp *watchParty) string {
	if len(wp.Viewers) == 0 {
		return "Nobody is in"
	}
	viewers := ""
	for _, v := range wp.Viewers {
		if len(v.Name) > 0 {
			viewers = viewers + v.Name + "\n"
			continue
		}
		viewers = viewers + "@" + v.Username + "\n"
	}
	return "The following are in: \n\n" + viewers
}

func getReadyMsg(wpID string) string {
	m := "Pause at 3 seconds\n\nReady status will last for 10 seconds."
	wp := getWatchPartyByID(wpID)
	if len(wp.Viewers) == 0 {
		return m
	}
	readyViewers := ""
	notReadyViewers := ""
	for _, v := range wp.Viewers {
		if v.Ready {
			readyViewers = readyViewers + getViewerName(v) + "\n"
			continue
		}
		notReadyViewers = notReadyViewers + getViewerName(v) + "\n"
	}
	return m + "\n\nNot Ready:\n\n" + notReadyViewers + "\nReady:\n\n" + readyViewers
}

func setViewerStatus(wpID string, vID int, ready bool, m *tb.Message, b *tb.Bot, readyNotReady *tb.ReplyMarkup) {
	wp := getWatchPartyByID(wpID)
	for _, vw := range wp.Viewers {
		if vID == vw.ID {
			vw.Ready = ready
			if ready {
				vw.Timer = revertViewerStatusAfter15Seconds(wpID, vID, m, b, readyNotReady)
			}
			//vw.Timer.Stop()
		}
	}
	b.Edit(m, getReadyMsg(wpID), readyNotReady)
}

func revertViewerStatusAfter15Seconds(wpID string, vID int, m *tb.Message, b *tb.Bot, readyNotReady *tb.ReplyMarkup) *time.Timer {
	timer := time.AfterFunc(time.Second * 10, func() {
		setViewerStatus(wpID, vID, false, m, b, readyNotReady)
	})
	return timer
}

func addPersonToWP(wp *watchParty, name string, username string, id int) {
	v := &viewer{ID: id, Name: name, Username: username}
	if !viewerExists(wp, v) {
		wp.Viewers = append(wp.Viewers, v)
	}
}

func viewerExists(wp *watchParty, v *viewer) bool {
	for _, vw := range wp.Viewers {
		if v.ID == vw.ID {
			return true
		}
	}
	return false
}

func removeViewerFromWP(wp *watchParty, v *viewer) {
	for i, vw := range wp.Viewers {
		if v.ID == vw.ID {
			wp.Viewers = append(wp.Viewers[:i], wp.Viewers[i+1:]...)
		}
	}
}
