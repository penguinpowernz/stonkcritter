package politstonk

import "github.com/timshannon/badgerhold/v4"

type Rep struct {
	Name string
}

func (bot *Bot) StoreReps(ds []Disclosure) {
	for _, d := range ds {
		bot.store.Insert(d.Representative, Rep{Name: d.Representative})
	}
}

func (bot *Bot) searchReps(search string) ([]string, error) {
	reps := []Rep{}
	err := bot.store.Find(
		&reps,
		badgerhold.Where("Name").Contains(search),
	)

	s := []string{}
	for _, r := range reps {
		s = append(s, r.Name)
	}

	return s, err
}
