package sinks

import (
	"github.com/nats-io/nats.go"
	"github.com/penguinpowernz/stonkcritter/models"
)

// NATSMessage will send the formatted disclosure message over NATS
func NATSMessage(nc *nats.Conn, subj string) Sink {
	return func(d models.Disclosure) error {
		err := nc.Publish(subj, []byte(d.String()))
		logerr(err, "natsmsg")
		return err
	}
}

// NATS will send the full disclosure object as JSON over NATS
func NATS(nc *nats.Conn, subj string) Sink {
	return func(d models.Disclosure) error {
		err := nc.Publish(subj, d.Bytes())
		logerr(err, "nats")
		return err
	}
}
