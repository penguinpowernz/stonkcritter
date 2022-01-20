package models

import (
	"testing"
)

func TestDisclosureTopic(t *testing.T) {
	d := Disclosure{Ticker: "MSFT", Senator: "Bob"}
	if d.CritterTopic() != "Bob" {
		t.FailNow()
	}

	if d.TickerTopic() != "$MSFT" {
		t.FailNow()
	}

	d = Disclosure{Ticker: "MSFT", Representative: "Bob"}
	if d.CritterTopic() != "Bob" {
		t.FailNow()
	}

	if d.TickerTopic() != "$MSFT" {
		t.FailNow()
	}
}
