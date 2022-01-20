package sinks

import (
	"fmt"
	"io"

	"github.com/penguinpowernz/stonkcritter/models"
)

func Writer(w io.Writer) Sink {
	return func(d models.Disclosure) error {
		fmt.Fprintln(w, d.CritterName(), d.Type, d.TickerString(), d.Amount, d.DaysAgo(), "days ago", d.AssetDescription, d.OwnerString())
		return nil
	}
}
