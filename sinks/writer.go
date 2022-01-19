package sinks

import (
	"fmt"
	"io"

	"github.com/penguinpowernz/stonkcritter/models"
)

func Writer(w io.Writer) Sink {
	return func(d models.Disclosure) error {
		fmt.Fprintln(w, d.String())
		return nil
	}
}
