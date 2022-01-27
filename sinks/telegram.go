package sinks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/penguinpowernz/stonkcritter/models"
	"golang.org/x/time/rate"
	tb "gopkg.in/tucnak/telebot.v2"
)

// TelegramChannel will create a sink that sends the formatted disclosure message
// to the given telegram channel using the given token
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
	lock := new(sync.Mutex)

	return func(d models.Disclosure) error {
		lock.Lock()
		defer lock.Unlock()
		channelLimit.Wait(context.Background())

		msg := tgEscape(d.String())
		_, err := b.Send(tb.ChatID(bcChan), msg, tb.ModeMarkdownV2, tb.NoPreview)
		logerr(err, "tgchannel", "sending broading message")
		if err == nil {
			Counts.TelegramChannel++
		}
		return err
	}, nil
}

type SubSender interface {
	Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error)
	Subs() []models.Sub
}

// TelegramBot will create a sink that uses the subscriptions contained in the provided bot to
// send direct messages to subscribed users containing the formatted disclosure message
func TelegramBot(bot SubSender) Sink {
	dmLimit := *rate.NewLimiter(rate.Every(time.Minute/59), 1)
	lock := new(sync.Mutex)

	// create a deduplicator so we don't send the same message to the same user twice
	// e.g. if they are subscribed to Pelosi and $MSFT and Pelosi makes an $MSFT trade
	shouldSend, markSent := deduper()

	return func(d models.Disclosure) error {
		lock.Lock()
		defer lock.Unlock()

		for _, s := range bot.Subs() {
			if !s.ShouldNotify(d) { // check if this subscription is for this disclosure
				continue
			}

			msg := tgEscape(d.String())
			if !shouldSend(s.ChatID, msg) { // check if we already sent this to the user
				continue
			}

			dmLimit.Wait(context.Background())

			if _, err := bot.Send(tb.ChatID(s.ChatID), msg, tb.ModeMarkdownV2, tb.NoPreview); err != nil {
				logerr(err, "tgbot", "sending disclosure message")
			}

			markSent(s.ChatID, msg)
		}

		Counts.TelegramBot++
		return nil
	}
}

func tgEscape(s string) string {
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	s = strings.ReplaceAll(s, "-", `\-`)
	s = strings.ReplaceAll(s, "$", `\$`)
	s = strings.ReplaceAll(s, ":", `\:`)
	return s
}
