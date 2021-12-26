package politstonk

import (
	"testing"
)

func TestDisclosureAfter(t *testing.T) {
	dd := []Disclosure{
		{DisclosureDate: "12/29/2020"},
		{DisclosureDate: "12/13/2020"},
		{DisclosureDate: "10/13/2020"},
	}

	after := Disclosures(dd).After(Date{"2020-10-01"})
	if len(after) != 3 {
		t.Log("expected 3 disclosures but was", len(after))
		t.FailNow()
	}

	// don't include today
	after = Disclosures(dd).After(Date{"2020-10-13"})
	if len(after) != 2 {
		t.Log("expected 2 disclosures but was", len(after))
		t.FailNow()
	}

	after = Disclosures(dd).After(Date{"2020-12-01"})
	if len(after) != 2 {
		t.Log("expected 2 disclosures but was", len(after))
		t.FailNow()
	}

	after = Disclosures(dd).After(Date{"2020-12-29"})
	if len(after) != 0 {
		t.Log("expected 0 disclosures but was", len(after))
		t.FailNow()
	}

	after = Disclosures(dd).After(Date{"2020-12-30"})
	if len(after) != 0 {
		t.Log("expected 0 disclosures but was", len(after))
		t.FailNow()
	}

	after = Disclosures(dd).After(Date{"2021-12-30"})
	if len(after) != 0 {
		t.Log("expected 0 disclosures but was", len(after))
		t.FailNow()
	}
}

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
