package politstonk

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/timshannon/badgerhold/v4"
	"golang.org/x/time/rate"
	tb "gopkg.in/tucnak/telebot.v2"
)

func NewBot(brain *badgerhold.Store, token string, bcChannel string) (*Bot, error) {
	b, err := tb.NewBot(tb.Settings{Token: token})
	if err != nil {
		return nil, err
	}

	bcChan, err := strconv.ParseInt(bcChannel, 10, 64)
	if err != nil {
		fmt.Println("WARN: failed to parse BOT_CHANNEL:", err)
	}

	bot := &Bot{
		Bot:          b,
		bcChannel:    bcChan,
		store:        brain,
		channelLimit: *rate.NewLimiter(rate.Every(time.Minute/19), 1),
	}

	bot.setupCommands()
	return bot, nil
}

type Bot struct {
	*tb.Bot
	bcChannel    int64
	store        *badgerhold.Store
	LogOnly      bool
	channelLimit rate.Limiter
}

func (bot *Bot) ConsumeDisclosures(dd []Disclosure) {
	log.Printf("consuming %d disclosures", len(dd))
	log.Println("adding any unknown reps")
	bot.StoreReps(dd)

	log.Printf("bot cursor is currently %s", bot.GetCursor())
	dd = Disclosures(dd).After(bot.GetCursor())
	log.Printf("%d disclosures found after the cursor", len(dd))

	for _, d := range dd {
		if bot.LogOnly {
			log.Println(d.String())
			continue
		}

		bot.Broadcast(d.String())
		bot.DispatchDisclosure(d)
		time.Sleep(time.Second)
	}

	bot.UpdateCursor()
}

func (bot *Bot) DispatchDisclosure(d Disclosure) {
	bot.store.ForEach(&badgerhold.Query{}, func(s Sub) {
		if s.IsTickerSub() {
			if d.Ticker == s.Ticker() {
				bot.Send(tb.ChatID(s.ChatID), d.String())
			}
			return
		}

		if strings.Contains(d.Representative, s.Topic) {
			bot.Send(tb.ChatID(s.ChatID), d.String())
		}
	})
}

func (bot *Bot) Broadcast(msg string) {
	if bot.bcChannel == 0 {
		return
	}

	bot.channelLimit.Wait(context.Background())
	if _, err := bot.Send(tb.ChatID(bot.bcChannel), msg, tb.ModeMarkdownV2); err != nil {
		log.Println("MESSAGE", msg)
		log.Printf("ERROR: sending broadcast %s", err)
	}
}

func (bot *Bot) UpdateCursor() {
	// TODO: error checking!!!
	bot.store.Update("cursor", time.Now())
}

func (bot *Bot) GetCursor() Date {
	var d Date
	err := bot.store.Get("cursor", &d)
	if err == badgerhold.ErrNotFound {
		d := NewDate(time.Now())
		// TODO: error checking!!!
		bot.store.Insert("cursor", d)
	}
	return d
}

type Date struct{ S string }

// NewDate will return a new date object from the given time
func NewDate(t time.Time) Date {
	return Date{fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())}
}

func (d Date) Time() time.Time {
	t, _ := time.Parse("2006-01-02", string(d.S))
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}
