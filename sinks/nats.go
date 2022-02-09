package sinks

import (
	"github.com/nats-io/nats.go"
	"github.com/penguinpowernz/stonkcritter/models"
)

// NATS will send the disclosures over NATS
func NATS(nc *nats.Conn, subj string) Sink {
	return func(d models.Disclosure) error {
		err := nc.Publish(subj, NewPayload(d).Bytes())
		logerr(err, "nats")
		return err
	}
}
