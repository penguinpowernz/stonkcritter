package sinks

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/penguinpowernz/stonkcritter/models"
	"golang.org/x/time/rate"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TelegramChannel(botToken, botChannel string) (Sink, error) {
	bcChan, err := strconv.ParseInt(botChannel, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse channel ID:", err)
	}

	b, err := tb.NewBot(tb.Settings{Token: botToken})
	if err != nil {
		return nil, err
	}
	channelLimit := *rate.NewLimiter(rate.Every(time.Minute/19), 1)

	return func(d models.Disclosure) error {
		channelLimit.Wait(context.Background())
		_, err := b.Send(tb.ChatID(bcChan), d.String(), tb.ModeMarkdownV2, tb.NoPreview)
		return err
	}, nil
}

type SubSender interface {
	Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error)
	Subs() []models.Sub
}

func TelegramBot(bot SubSender) Sink {
	dmLimit := *rate.NewLimiter(rate.Every(time.Minute/59), 1)

	// create a deduplicator so we don't send the same message to the same user twice
	// e.g. if they are subscribed to Pelosi and $MSFT and Pelosi makes an $MSFT trade
	shouldSend, markSent := deduper()

	return func(d models.Disclosure) error {
		for _, s := range bot.Subs() {
			if !s.ShouldNotify(d) { // check if this subscription is for this disclosure
				continue
			}

			msg := d.String()
			if !shouldSend(s.ChatID, msg) { // check if we already sent this to the user
				continue
			}

			dmLimit.Wait(context.Background())

			if _, err := bot.Send(tb.ChatID(s.ChatID), msg, tb.ModeMarkdownV2, tb.NoPreview); err != nil {
				log.Println("ERROR: disaptching disclosure:", err)
			}

			markSent(s.ChatID, msg)
		}

		return nil
	}
}
