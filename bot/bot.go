package bot

import (
	"log"
	"time"

	"github.com/penguinpowernz/stonkcritter/models"
	"github.com/timshannon/badgerhold/v4"
	tb "gopkg.in/tucnak/telebot.v2"
)

func NewBot(brain *Brain, token string) (*Bot, error) {
	b, err := tb.NewBot(tb.Settings{Token: token, Poller: &tb.LongPoller{Timeout: 10 * time.Second}})
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Bot:   b,
		brain: brain,
	}

	bot.loadSubs()
	bot.setupCommands()
	go bot.Start()
	return bot, nil
}

type Bot struct {
	*tb.Bot
	brain   *Brain
	LogOnly bool
	subs    []models.Sub
}

func (bot *Bot) Subs() []models.Sub {
	return bot.subs
}

func (bot *Bot) loadSubs() {
	if err := bot.brain.Find(&bot.subs, &badgerhold.Query{}); err != nil {
		log.Println("ERROR: couldn't load subs:", err)
	}
}
