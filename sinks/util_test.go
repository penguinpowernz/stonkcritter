package sinks

import "testing"

func TestDeDuper(t *testing.T) {
	shouldsend, marksent := deduper()

	var id int64
	var msg string

	id = 44
	msg = "bob ya dingus"
	if !shouldsend(id, msg) {
		t.FailNow() // there is nothing in there yet, so should allow sending
	}
	marksent(id, msg)
	if shouldsend(id, msg) {
		t.FailNow() // we marked it as sent so it should disallow sending
	}

	id = 45 // should allow the same message to a separate ID
	msg = "bob ya dingus"
	if !shouldsend(id, msg) {
		t.FailNow() // there is nothing in there yet, so should allow sending
	}
	marksent(id, msg)
	if shouldsend(id, msg) {
		t.FailNow() // we marked it as sent so it should disallow sending
	}

	id = 45
	msg = "bob ya dingleberry" // should allow a different message to the same ID
	if !shouldsend(id, msg) {
		t.FailNow() // there is nothing in there yet, so should allow sending
	}
	marksent(id, msg)
	if shouldsend(id, msg) {
		t.FailNow() // we marked it as sent so it should disallow sending
	}
}
