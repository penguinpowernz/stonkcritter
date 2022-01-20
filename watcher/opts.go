package watcher

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/penguinpowernz/stonkcritter/source"
)

type Option func(w *Watcher) error

func DiskCursor(filename string, autosave bool) Option {
	return func(w *Watcher) error {
		data, err := ioutil.ReadFile(filename)

		if err != nil {
			return err
		}

		epoch, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		if err != nil {
			return err
		}
		w.cursor = time.Unix(epoch, 0)

		w.autosaveCursor = func(cursor time.Time) error {
			return ioutil.WriteFile(filename, []byte(strconv.Itoa(int(w.cursor.Unix()))), 0644)
		}

		return nil
	}
}

func StartAt(t time.Time) Option {
	return func(w *Watcher) error {
		w.cursor = t
		return nil
	}
}

func FromFile(filename string) Option {
	return func(w *Watcher) error {
		w.provider = source.GetDisclosuresFromFile(filename)
		return nil
	}
}

func FromS3() Option {
	return func(w *Watcher) error {
		w.provider = source.GetDisclosuresFromS3
		return nil
	}
}
