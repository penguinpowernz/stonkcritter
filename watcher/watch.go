package watcher

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/penguinpowernz/stonkcritter/models"
)

type Provider func() ([]models.Disclosure, error)

func NewWatcher(opts ...Option) (*Watcher, error) {
	w := new(Watcher)

	// set some sane defaults
	w.ticker = time.NewTicker(24 * time.Hour)
	w.cursor = time.Now().Add(time.Hour * 24 * 30 * -1) // default to only send disclosures from previous 30 days
	w.autosaveCursor = func(time.Time) error { return nil }
	w.critters = map[string]interface{}{}
	w.crittersLock = new(sync.RWMutex)
	w.onCheckDoneLock = new(sync.Mutex)

	for _, o := range opts {
		if err := o(w); err != nil {
			return nil, err
		}
	}

	if w.provider == nil {
		return nil, errors.New("must set a provider using FromS3 or FromFile options")
	}

	w.trades = make(chan models.Disclosure, 100)
	w.manualCheck = make(chan bool)
	return w, nil
}

type Watcher struct {
	ticker          *time.Ticker
	cursor          time.Time
	running         bool
	trades          chan models.Disclosure
	provider        Provider
	autosaveCursor  func(time.Time) error
	manualCheck     chan bool
	checks          int
	dispatched      int
	inflight        int
	onCheckDone     []chan struct{}
	onCheckDoneLock *sync.Mutex

	critters     map[string]interface{}
	crittersLock *sync.RWMutex
}

// Critters is a list of all the congress critters known to make trades
func (w Watcher) Critters() []string {
	w.crittersLock.RLock()
	defer w.crittersLock.RUnlock()

	var cs []string
	for k := range w.critters {
		cs = append(cs, k)
	}
	return cs
}

// Checks returns the number of disclosure source checks that have been done
func (w Watcher) Checks() int {
	return w.checks
}

func (w Watcher) Dispatched() int {
	return w.dispatched
}

// Inflight tells how many disclosures are still waiting to be pulled from
// the watcher (via the `Disclosure`` method)
func (w Watcher) Inflight() int {
	return w.inflight
}

func (w Watcher) CurrentCursor() time.Time {
	return w.cursor
}

func (w *Watcher) WaitForCheck() {
	w.onCheckDoneLock.Lock()
	defer w.onCheckDoneLock.Unlock()
	ch := make(chan struct{})
	w.onCheckDone = append(w.onCheckDone, ch)
	<-ch
}

// reset all the check waiters
func (w *Watcher) checked() {
	w.onCheckDoneLock.Lock()
	defer w.onCheckDoneLock.Unlock()
	for _, ch := range w.onCheckDone {
		close(ch)
	}
	w.onCheckDone = []chan struct{}{}
}

// CheckNow will trigger the watcher to check the disclosers from
// the provider immediately
func (w Watcher) CheckNow() {
	w.manualCheck <- true
}

func (w *Watcher) Start(ctx context.Context) {
	w.running = true

	var getDisclosures = func() {
		expo := backoff.NewExponentialBackOff()
		var dd []models.Disclosure

		backoff.Retry(func() (err error) {
			dd, err = w.provider()
			return err
		}, expo)

		w.dispatch(dd)
		w.checked()
	}

	go func() {
		defer func() { w.running = false }()

		getDisclosures()

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.manualCheck:
				getDisclosures()
			case <-w.ticker.C:
				getDisclosures()
			}
		}
	}()
}

func (w *Watcher) dispatch(trades []models.Disclosure) {
	w.crittersLock.Lock()
	defer w.crittersLock.Unlock()

	log.Printf("checking %d trade disclosures, current cursor is %s", len(trades), w.cursor.Format("2006-01-02"))

	var dispatched int
	for _, t := range trades {
		w.critters[t.CritterName()] = nil

		if t.DisclosedOn().Before(w.cursor) {
			continue
		}

		w.trades <- t
		w.cursor = t.DisclosedOn()
		dispatched++
		w.inflight++

		if err := w.autosaveCursor(w.cursor); err != nil {
			log.Println("ERROR: autosaving cursor")
		}
	}

	log.Printf("found %d new trade disclosures, cursor is now at %s", dispatched, w.cursor.Format("2006-01-02"))
	w.checks++
	w.dispatched += dispatched
}

// Next will block until there are available disclosures, and then return true, or
// if the watcher is no longer running return false.  This can be used for iterating
// in for loops and getting the next disclosure using Disclosure()
func (w *Watcher) Next() bool {
	for {
		if !w.running {
			return false
		}

		if len(w.trades) == 0 {
			time.Sleep(time.Second)
			continue
		}

		return true
	}
}

func (w *Watcher) Disclosure() models.Disclosure {
	w.inflight--
	return <-w.trades
}
