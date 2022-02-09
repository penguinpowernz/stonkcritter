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
	D   models.Disclosure `json:"raw"`
	C   string            `json:"critter"`
	AT  string            `json:"asset_type"`
	TT  string            `json:"ticker"`
	TRT string            `json:"type"`
	OB  string            `json:"owner"`
	AE  string            `json:"amount_emojis"`
	TE  string            `json:"type_emojis"`
	ID  string            `json:"id"`
	TD  string            `json:"transaction_date"`
	DD  string            `json:"disclosure_date"`
	PDF bool              `json:"is_pdf_disclosure"`
	S   string            `json:"formatted_string"`
}

func NewPayload(dis models.Disclosure) Payload {
	return Payload{
		D:   dis,
		C:   dis.CritterName(),
		AT:  dis.AssetTypeTopic(),
		TT:  dis.TickerTopic(),
		TRT: dis.TradeType(),
		OB:  dis.OwnerString(),
		AE:  dis.AmountEmojis(),
		TE:  dis.TypeEmoji(),
		ID:  dis.ID(),
		TD:  dis.TransactionOn().Format("2006-01-02"),
		DD:  dis.DisclosedOn().Format("2006-01-02"),
		PDF: dis.IsPDFDisclosedFiling(),
		S:   dis.String(),
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
