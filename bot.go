package stonkcritter

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/timshannon/badgerhold/v4"
	"golang.org/x/time/rate"
	tb "gopkg.in/tucnak/telebot.v2"
)

func NewBot(brain *badgerhold.Store, token string, bcChannel string) (*Bot, error) {
	b, err := tb.NewBot(tb.Settings{Token: token, Poller: &tb.LongPoller{Timeout: 10 * time.Second}})
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
		dmLimit:      *rate.NewLimiter(rate.Every(time.Minute/59), 1),
	}

	bot.setupCommands()
	go bot.Start()
	return bot, nil
}

type Bot struct {
	*tb.Bot
	bcChannel    int64
	store        *badgerhold.Store
	LogOnly      bool
	channelLimit rate.Limiter
	dmLimit      rate.Limiter
}

func (bot *Bot) ConsumeDisclosures(dd []Disclosure) {
	log.Printf("consuming %d disclosures", len(dd))
	log.Println("adding any unknown reps")
	bot.StoreCritters(dd)

	log.Printf("bot cursor is currently %s", bot.GetCursor())
	dd = Disclosures(dd).After(bot.GetCursor())
	log.Printf("%d disclosures found after the cursor", len(dd))

	var subs []Sub
	if err := bot.store.Find(&subs, &badgerhold.Query{}); err != nil {
		log.Println("ERROR: couldn't load subs:", err)
		return
	}

	log.Println(len(subs), "subs loaded, ready to dispatch")

	for _, d := range dd {
		if bot.LogOnly {
			log.Println(d.String())
			continue
		}

		bot.Broadcast(d.String())
		bot.DispatchDisclosure(d, subs)
		time.Sleep(time.Second)
	}

	bot.UpdateCursor()
}

func (bot *Bot) DispatchDisclosure(d Disclosure, subs []Sub) {
	// create a deduplicator so we don't send the same message to the same user twice
	// e.g. if they are subscribed to Pelosi and $MSFT and Pelosi makes an $MSFT trade
	shouldSend, markSent := deduper()

	for _, s := range subs {
		if !s.ShouldNotify(d) { // check if this subscription is for this disclosure
			continue
		}

		msg := d.String()
		if !shouldSend(s.ChatID, msg) { // check if we already sent this to the user
			continue
		}

		bot.dmLimit.Wait(context.Background())

		if _, err := bot.Send(tb.ChatID(s.ChatID), msg, tb.ModeMarkdownV2, tb.NoPreview); err != nil {
			log.Println("ERROR: disaptching disclosure:", err)
		}

		markSent(s.ChatID, msg)
	}
}

func (bot *Bot) Broadcast(msg string) {
	if bot.bcChannel == 0 {
		return
	}

	bot.channelLimit.Wait(context.Background())
	if _, err := bot.Send(tb.ChatID(bot.bcChannel), msg, tb.ModeMarkdownV2, tb.NoPreview); err != nil {
		log.Println("MESSAGE", msg)
		log.Printf("ERROR: sending broadcast %s", err)
	}
}

func (bot *Bot) UpdateCursor() {
	log.Println("updating cursor to", time.Now().Format("2006-01-02"))
	if err := bot.store.Upsert("cursor", NewDate(time.Now())); err != nil {
		log.Println("ERROR: failed to update the cursor!", err)
	}
}

func (bot *Bot) GetCursor() Date {
	var d Date
	err := bot.store.Get("cursor", &d)
	if err == badgerhold.ErrNotFound {
		log.Println("WARNING: no cursor found, creating new one for", time.Now().Format("2006-01-02"))
		d := NewDate(time.Now())
		if err := bot.store.Insert("cursor", d); err != nil {
			log.Println("ERROR: failed to insert the cursor!", err)
		}
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
