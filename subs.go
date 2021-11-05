package politstonk

import (
	"fmt"
	"strings"
)

type Sub struct {
	ChatID int32
	Topic  string
}

func (s Sub) String() string {
	return fmt.Sprintf("%d.%s", s.ChatID, strings.ReplaceAll(s.Topic, " ", "_"))
}

func (s Sub) IsTickerSub() bool {
	return isTicker(s.Topic)
}

func (s Sub) Ticker() string {
	return strings.ReplaceAll(s.Topic, "$", "")
}
