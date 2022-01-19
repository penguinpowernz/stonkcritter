package bot

import (
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/penguinpowernz/stonkcritter/models"
	"github.com/timshannon/badgerhold/v4"
)

func NewBrain(dataDir string) (*Brain, error) {
	opts := badgerhold.DefaultOptions
	opts.Options = badger.DefaultOptions(dataDir)
	st, err := badgerhold.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Brain{st}, nil
}

type Brain struct {
	*badgerhold.Store
}

func (br *Brain) ListCritters() ([]string, error) {
	cs := []models.Critter{}
	err := br.Find(&cs, nil)
	ss := []string{}
	for _, s := range cs {
		ss = append(ss, s.Name)
	}
	return ss, err
}

func (br *Brain) StoreCritters(names []string) {
	for _, n := range names {
		br.Insert(n, models.Critter{Name: n})
	}
}

func (br *Brain) SearchCritters(search string) ([]string, error) {
	cs := []models.Critter{}

	var q *badgerhold.Query
	if search == "" {
		q = new(badgerhold.Query)
	} else {
		q = badgerhold.Where("Name").MatchFunc(func(ra *badgerhold.RecordAccess) (bool, error) {
			n := ra.Record().(*models.Critter).Name
			n = strings.ToLower(n)
			return strings.Contains(n, strings.ToLower(search)), nil
		})
	}

	err := br.Find(
		&cs,
		q,
	)

	s := []string{}
	for _, c := range cs {
		s = append(s, c.Name)
	}

	return s, err
}

func (br *Brain) Subscribe(id int64, topic string) error {
	s := models.Sub{ChatID: id, Topic: topic}
	return br.Insert(s.String(), s)
}

func (br *Brain) Unsubscribe(id int64, topic string) error {
	s := models.Sub{ChatID: id, Topic: topic}
	return br.Delete(s.String(), s)
}
