package sinks

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/penguinpowernz/stonkcritter/models"
)

func Webhook(url string) Sink {
	return func(d models.Disclosure) error {
		r, err := http.Post(url, "application/json", bytes.NewReader(d.Bytes()))
		if err != nil {
			logit("webhook", "ERROR: failed to complete request: %s", err)
			return err
		}

		if r.StatusCode != 200 {
			err := errors.New("unexpected status code: " + r.Status)
			logit("webhook", "ERROR: %s", err)
			return err
		}

		Counts.Webhook++
		return nil
	}
}
