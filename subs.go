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
