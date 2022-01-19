package sinks

import (
	"github.com/nats-io/nats.go"
	"github.com/penguinpowernz/stonkcritter/models"
)

type Sink func(models.Disclosure) error

// NATSMessage will send the formatted disclosure message over NATS
func NATSMessage(nc nats.Conn, subj string) Sink {
	return func(d models.Disclosure) error {
		return nc.Publish(subj, []byte(d.String()))
	}
}

// NATS will send the full disclosure object as JSON over NATS
func NATS(nc nats.Conn, subj string) Sink {
	return func(d models.Disclosure) error {
		return nc.Publish(subj, d.Bytes())
	}
}
