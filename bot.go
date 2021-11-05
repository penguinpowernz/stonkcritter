package politstonk

import (
	"strings"

	"github.com/timshannon/badgerhold/v4"
	tb "gopkg.in/tucnak/telebot.v2"
)

func NewBot(token string, bcChannel int32) (*Bot, error) {
	b, err := tb.NewBot(tb.Settings{Token: token})
	if err != nil {
		return nil, err
	}
	bot := &Bot{Bot: b, bcChannel: bcChannel}
	bot.setupCommands()
	return bot, nil
}

type Bot struct {
	*tb.Bot
	bcChannel int32
	store     *badgerhold.Store
}

func (bot *Bot) HandleDisclosure(d Disclosure) {
	bot.store.ForEach(&badgerhold.Query{}, func(s Sub) {
		if len(s.Topic) < 6 {
			if s.Topic == d.Ticker {
				bot.Send(tb.ChatID(s.ChatID), d.String())
			}
		} else {
			if strings.Contains(d.Representative, s.Topic) {
				bot.Send(tb.ChatID(s.ChatID), d.String())
			}
		}
	})
}

func (bot *Bot) Broadcast(msg string) {
	if bot.bcChannel == 0 {
		return
	}
	bot.Send(tb.ChatID(bot.bcChannel), msg)
}
