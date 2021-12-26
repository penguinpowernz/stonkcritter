package politstonk

import (
	"fmt"
	"strings"
)

type Sub struct {
	ChatID int64
	Topic  string
}

func (s Sub) String() string {
	return fmt.Sprintf("%d.%s", s.ChatID, strings.ReplaceAll(s.Topic, " ", "_"))
}

func (s Sub) ShouldNotify(d Disclosure) bool {
	switch s.Topic {
	case d.TickerTopic():
	case d.AssetTypeTopic():
	case d.CritterTopic():
	default:
		return false
	}
	return true
}
