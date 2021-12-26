package politstonk

import "testing"

func TestSubShouldNotify(t *testing.T) {
	s := Sub{1, "$MSFT"}
	d := Disclosure{Ticker: "MSFT", Senator: "Bob"}
	d1 := Disclosure{Ticker: "MSFT", Senator: "Frank"}

	if !s.ShouldNotify(d) {
		t.FailNow()
	}

	if !s.ShouldNotify(d1) {
		t.FailNow()
	}

	s1 := Sub{1, "Bob"}
	d2 := Disclosure{Ticker: "MSFT", Senator: "Bob"}
	d3 := Disclosure{Ticker: "MSFT", Senator: "Bob"}

	if !s1.ShouldNotify(d2) {
		t.FailNow()
	}

	if !s1.ShouldNotify(d3) {
		t.FailNow()
	}
}
