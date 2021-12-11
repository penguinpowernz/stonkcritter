package politstonk

import (
	"strings"

	"github.com/timshannon/badgerhold/v4"
)

type Critter struct {
	Name string
}

func (bot *Bot) StoreCritters(ds []Disclosure) {
	for _, d := range ds {
		if d.Representative != "" {
			bot.store.Insert(d.Representative, Critter{Name: d.Representative})
		}

		if d.Senator != "" {
			bot.store.Insert(d.Senator, Critter{Name: d.Senator})
		}
	}
}

func (bot *Bot) searchCritters(search string) ([]string, error) {
	cs := []Critter{}

	var q *badgerhold.Query
	if search == "" {
		q = new(badgerhold.Query)
	} else {
		q = badgerhold.Where("Name").MatchFunc(func(ra *badgerhold.RecordAccess) (bool, error) {
			n := ra.Record().(*Critter).Name
			n = strings.ToLower(n)
			return strings.Contains(n, strings.ToLower(search)), nil
		})
	}

	err := bot.store.Find(
		&cs,
		q,
	)

	s := []string{}
	for _, c := range cs {
		s = append(s, c.Name)
	}

	return s, err
}
