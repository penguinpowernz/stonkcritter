package sinks

import (
	"github.com/nats-io/nats.go"
	"github.com/penguinpowernz/stonkcritter/models"
)

type Sink func(models.Disclosure) error

func NATSMessage(nc nats.Conn, subj string) Sink {
	return func(d models.Disclosure) error {
		return nc.Publish(subj, []byte(d.String()))
	}
}
