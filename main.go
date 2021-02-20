package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	// "strconv"
	"github.com/pborman/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

type viewer struct {
	Name          string
	Username      string
	ID            int
	ReadyTimeLeft int
}

type watchParty struct {
	ID              string
	Name            string
	Viewers         []*viewer
	OwnerID         int
	EveryoneIsReady chan bool
	Ticker          *time.Ticker
	TickerRunning   bool
}

type replyID struct {
	ChatID int64
	MsgID  int
}

var data []*watchParty
var replyIDs []*replyID

const countdownDuration = 15

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
		ParseMode: tb.ModeMarkdownV2,
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	replyquery := &tb.ReplyMarkup{ForceReply: true, Selective: true}

	b.Handle("/watch", func(m *tb.Message) {
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

	b.Handle("/count", func(m *tb.Message) {
		sendCountdown(b, m.Chat)
	})

	b.Start()
}

func sendCountdown(b *tb.Bot, chat *tb.Chat) {
	b.Send(chat, "3")
	time.Sleep(1 * time.Second)
	b.Send(chat, "2")
	time.Sleep(1 * time.Second)
	b.Send(chat, "1")
	time.Sleep(1 * time.Second)
	b.Send(chat, "Go!")
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
	btnIn := InOrOut.Data("I'm in ✅", "in", wpID)
	btnOut := InOrOut.Data("I'm not in ❌", "out", wpID)
	btnInitiate := InOrOut.Data("Start countdown ▶️", "initiate", wpID)
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
		btnReady := readyNotReady.Data("Ready 🎬", "ready", wpID)
		btnNotReady := readyNotReady.Data("Not ready ❌", "notready", wpID)
		readyNotReady.Inline(InOrOut.Row(btnReady, btnNotReady))

		wp := getWatchPartyByID(wpID)
		mr, _ := b.Send(chat, getReadyMsg(wp), readyNotReady)
		startMainTicker(b, mr, wp, readyNotReady)

		b.Handle(&btnReady, func(c *tb.Callback) {
			b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are ready!"})
			wp := getWatchPartyByID(wpID)
			addPersonToWP(wp, c.Sender.FirstName, c.Sender.Username, c.Sender.ID)
			setViewerTimeRemaining(wp, c.Sender.ID, countdownDuration)
			b.Edit(mr, getReadyMsg(wp), readyNotReady)
			if checkIfWeAreAGo(c.Data) {
				wp.Ticker.Stop()
				wp.EveryoneIsReady <- true
				b.Edit(mr, getReadyMsg(wp), readyNotReady)
				go func() {
					b.Send(m.Chat, "Looks like we are all ready! Starting count.")
					time.Sleep(2 * time.Second)
					sendCountdown(b, m.Chat)
				}()
			}
		})
		b.Handle(&btnNotReady, func(c *tb.Callback) {
			b.Respond(c, &tb.CallbackResponse{Text: "Noted that you are not ready"})
			wp := getWatchPartyByID(c.Data)
			setViewerTimeRemaining(wp, c.Sender.ID, 0)
			b.Edit(mr, getReadyMsg(wp), readyNotReady)
		})
	})
}

func startMainTicker(b *tb.Bot, m *tb.Message, wp *watchParty, readyNotReady *tb.ReplyMarkup) {

	if !wp.TickerRunning {
		wp.Ticker = time.NewTicker(1 * time.Second)
		wp.EveryoneIsReady = make(chan bool)
		wp.TickerRunning = true

		go func() {
			for {
				select {
				case <-wp.EveryoneIsReady:
					wp.TickerRunning = false
					return
				case <-wp.Ticker.C:
					someoneTimedOut := updateViewerTimeRemaining(wp)
					if someoneTimedOut {
						b.Edit(m, getReadyMsg(wp), readyNotReady)
					}
				}
			}
		}()
	}
}

func checkIfWeAreAGo(wpID string) bool {
	wp := getWatchPartyByID(wpID)
	for _, v := range wp.Viewers {
		if v.ReadyTimeLeft == 0 {
			return false
		}
	}
	return true
}

func getViewerName(v *viewer) string {
	if len(v.Name) > 0 {
		return v.Name
	} else {
		return "@" + v.Username
	}
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
	return "*The following are in:* \n\n" + viewers
}

func getReadyMsg(wp *watchParty) string {
	m := "Get paused!\n\nReady status will last for " + strconv.Itoa(countdownDuration) + " seconds."
	if len(wp.Viewers) == 0 {
		return m
	}
	readyViewers := ""
	notReadyViewers := ""
	for _, v := range wp.Viewers {
		if v.ReadyTimeLeft > 0 {
			readyViewers = readyViewers + getViewerName(v) + "\n"
			continue
		}
		notReadyViewers = notReadyViewers + getViewerName(v) + "\n"
	}
	return m + "\n\nNot Ready:\n\n" + notReadyViewers + "\nReady:\n\n" + readyViewers
}

func setViewerTimeRemaining(wp *watchParty, vID int, timeRemaining int) {
	for _, vw := range wp.Viewers {
		if vID == vw.ID {
			vw.ReadyTimeLeft = timeRemaining
			return
		}
	}
}

func updateViewerTimeRemaining(wp *watchParty) (someoneTimedOut bool) {
	for _, vw := range wp.Viewers {
			if vw.ReadyTimeLeft > 0 {
				vw.ReadyTimeLeft--
				if vw.ReadyTimeLeft == 0 {
					someoneTimedOut = true
				}
			}
	}
	return someoneTimedOut
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
