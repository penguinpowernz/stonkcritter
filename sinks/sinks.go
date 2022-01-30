package sinks

import (
	"encoding/json"
	"log"
	"os"

	"github.com/penguinpowernz/stonkcritter/models"
)

var Counts = new(Stats)

type Stats struct {
	Websocket       int
	MQTT            int
	NATS            int
	Webhook         int
	TelegramChannel int
	TelegramBot     int
	Writer          int
}

type Payload struct {
	Formatted string            `json:"formatted"`
	Raw       models.Disclosure `json:"raw"`
}

func NewPayload(d models.Disclosure) Payload {
	return Payload{
		Formatted: d.String(),
		Raw:       d,
	}
}

func (pl Payload) Bytes() []byte {
	data, _ := json.Marshal(pl)
	return data
}

var logger = log.New(os.Stderr, "", log.Flags())

func logit(name string, msg string, args ...interface{}) {
	args = append([]interface{}{name}, args...)
	logger.Printf("[sink:%s] "+msg, args...)
}

func logerr(err error, name string, msgs ...string) {
	if err == nil {
		return
	}

	msg := ""
	if len(msgs) > 0 {
		msg = msgs[0] + ": "
	}
	logger.Printf("[sink:%s] ERROR: %s%s", name, msg, err)
}

type Sink func(models.Disclosure) error
