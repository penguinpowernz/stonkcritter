package politstonk

import (
	"encoding/base64"
	"log"
	"strings"

	"github.com/timshannon/badgerhold/v4"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *Bot) setupCommands() {
	bot.Bot.Handle("/critter", bot.findrep)
	bot.Bot.Handle("/list", bot.list)
	bot.Bot.Handle("/unfollow", bot.unfollow)
	bot.Bot.Handle("/follow", bot.follow)
	bot.Bot.Handle("/help", bot.help)
	bot.Bot.Handle("/start", bot.start)
	bot.Bot.Handle(tb.OnText, bot.kitchenSink)
}

func (bot *Bot) kitchenSink(msg *tb.Message) {
	switch {
	case strings.Contains(msg.Text, "/unfollow_"):
		bits := strings.Split(msg.Text, "_")
		if len(bits) != 2 {
			bot.Send(msg.Chat, "⚠️ Can't unfollow nothing...")
			return
		}

		data, err := base64.RawStdEncoding.DecodeString(bits[1])
		if err != nil {
			bot.Send(msg.Chat, "⚠️ Sorry, couldn't decode the thing that you want to unfollow...")
			return
		}

		thing := string(data)
		msg.Payload = thing
		bot.unfollow(msg)
	}
}

func (bot *Bot) start(msg *tb.Message) {
	bot.Send(msg.Chat, "Hey, welcome 👋 I'm going to run the /help command for you so you can see what I can do...")
	bot.help(msg)
}

func (bot *Bot) helpFollow(msg *tb.Message) {
	m := strings.Join([]string{
		"You can follow these types of things:",
		"",
		"A congress critter, e.g. `Nancy Pelosi` or `Ron Wyden`",
		"A stock ticker e.g. `$MSFT` or `$AAPL`",
		// "",
		// "The following asset types: ",
		// "  `#crypto` - Cryptocurrency",
		// "  `#comfuture` - Commodies/Futures Contract",
		// "  `#nopstock` - Non-Public Stock",
		// "  `#opts` - Stock Options",
		// "  `#corpbond` - Corporate Bonds",
		// "  `#osec` - Other Securities",
		// "  `#msec` - Municiple Securities",
		// "  `#pdf` - PDF Disclosed Filing",
	}, "\n")

	bot.Send(msg.Sender, m, tb.ModeMarkdown)
}

func (bot *Bot) follow(msg *tb.Message) {
	topic := msg.Payload

	if topic == "" {
		bot.helpFollow(msg)
		return
	}

	if !isTicker(topic) {
		bot.Send(msg.Chat, "let me see if I can find that congress critter in my list...")
		names, err := bot.searchCritters(topic)

		if err != nil {
			bot.Send(msg.Chat, "😔 sorry, failed to search: "+err.Error())
			return
		}

		if len(names) > 10 {
			bot.Send(msg.Chat, "sorry, I found more than 10 results with the same name, you'll need to be more specific, try /findrep")
			return
		}

		if len(names) > 1 {
			bot.Send(msg.Chat, "OK, I found 10 reps that it could be, can you be more specific?:\n"+strings.Join(names, "\n"))
			return
		}

		bot.Send(msg.Chat, "Cool, found "+names[0]+" in my list, using that")
		topic = names[0]
	}

	s := Sub{ChatID: int32(msg.Chat.ID), Topic: topic}
	err := bot.store.Insert(s.String(), s)
	if err != nil {
		bot.Send(msg.Chat, "😔 sorry, failed to save that: "+err.Error())
		return
	}

	bot.Send(msg.Chat, "OK saved 👍")
}

func (bot *Bot) unfollow(msg *tb.Message) {
	bot.Send(msg.Chat, "Looking for "+msg.Payload+" in your list of subscriptions...")
	s := Sub{ChatID: msg.Chat.ID, Topic: msg.Payload}
	err := bot.store.Delete(s.String(), s)
	if err != nil {
		bot.Send(msg.Chat, "😔 sorry, failed to delete that: "+err.Error())
		return
	}

	bot.Send(msg.Chat, "OK unfollowed 👍")
}

func (bot *Bot) list(msg *tb.Message) {
	subs := []Sub{}
	err := bot.store.Find(&subs, badgerhold.Where("ChatID").Eq(msg.Chat.ID))

	if err != nil {
		bot.Send(msg.Chat, "😔 sorry, failed to search: "+err.Error())
		return
	}

	if len(subs) == 0 {
		bot.Send(msg.Chat, "Not following anything, try /follow")
		return
	}

	topics := []string{}
	for _, s := range subs {
		topics = append(topics, "- "+s.Topic+" /unfollow_"+base64.RawStdEncoding.EncodeToString([]byte(s.Topic)))
	}

	m := "OK, you're following these things, click the unfollow link to remove them from the list:\n" + strings.Join(topics, "\n")
	bot.Send(msg.Chat, m)
}

func (bot *Bot) findrep(msg *tb.Message) {
	search := msg.Payload

	if len(search) < 4 {
		bot.Send(msg.Chat, "sorry, you're gonna need to do more than 3 letters")
		return
	}

	names, err := bot.searchCritters(search)

	if err != nil {
		bot.Send(msg.Chat, "sorry, failed to search: "+err.Error())
		return
	}

	if len(names) == 0 {
		bot.Send(msg.Chat, "sorry, couldn't find anyone with that name")
		return
	}

	if len(names) > 10 {
		bot.Send(msg.Chat, "sorry, too many results.  Please be more specific")
		return
	}

	if len(names) == 1 {
		bot.Send(msg.Chat, "Found them: "+names[0])
		return
	}

	if _, err := bot.Send(msg.Chat, "OK, it's gotta be one of these:\n"+strings.Join(names, "\n")); err != nil {
		log.Println("ERROR:", err)
	}
}

func (bot *Bot) help(msg *tb.Message) {
	txt := []string{}
	txt = append(txt, "This bot will help you keep track of the stocks that the US congress critters trade.  Bear in mind that they have 45 days to disclose their trades, and most have started leaving this to the last day possible due to public srutiny of their trades.")
	txt = append(txt, "")
	txt = append(txt, "/help - `this help text`")
	txt = append(txt, "/follow <thing> - `the main function, allows you to follow a $TICKER or a congress critter by name.  The name maybe converted to the full name as it appears in the disclosure source`")
	txt = append(txt, "/unfollow <thing> - `unfollow something from your list`")
	txt = append(txt, "/list - `this will list everything that you are following`")
	txt = append(txt, "/critter <name> - `if you're having trouble finding a congress critter by name, you can use this to search for the name you should use in the follow command`")

	if _, err := bot.Send(msg.Chat, strings.Join(txt, "\n"), tb.ModeMarkdown); err != nil {
		log.Println("ERROR:", err)
	}
}
