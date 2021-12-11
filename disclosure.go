package politstonk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

var DisclosuresURLHouse = "https://house-stock-watcher-data.s3-us-west-2.amazonaws.com/data/all_transactions.json"
var DisclosuresURLSenate = "https://senate-stock-watcher-data.s3-us-west-2.amazonaws.com/aggregate/all_transactions.json"

// {
// 	"disclosure_year": 2021,
// 	"disclosure_date": "10/04/2021",    DD/MM/YYYY
// 	"transaction_date": "2021-09-27",   YYYY-MM-DD
// 	"owner": "joint",
// 	"ticker": "BP",
// 	"asset_description": "BP plc",
// 	"type": "purchase",
// 	"amount": "$1,001 - $15,000",
// 	"representative": "Hon. Virginia Foxx",
// 	"district": "NC05",
// 	"ptr_link": "https://disclosures-clerk.house.gov/public_disc/ptr-pdfs/2021/20019557.pdf",
// 	"cap_gains_over_200_usd": false
// }

type Disclosure struct {
	DisclosureYear     int    `json:"disclosure_year"`
	DisclosureDate     string `json:"disclosure_date"`
	TransactionDate    string `json:"transaction_date"`
	Owner              string `json:"owner"`
	Ticker             string `json:"ticker"`
	AssetDescription   string `json:"asset_description"`
	Type               string `json:"type"`
	Amount             string `json:"amount"`
	Representative     string `json:"representative"`
	District           string `json:"district"`
	PtrLink            string `json:"ptr_link"`
	CapGainsOver200Usd bool   `json:"cap_gains_over_200_usd"`
}

func (dis Disclosure) AmountEmojis() string {
	var c int
	switch dis.Amount {
	case "$50,000,000 +":
		c = 9
	case "$5,000,001 - $25,000,000":
		c = 8
	case "$1,000,001 - $5,000,000", "$1,000,000 +":
		c = 7
	case "$500,001 - $1,000,000":
		c = 6
	case "$250,001 - $500,000":
		c = 5
	case "$100,001 - $250,000":
		c = 4
	case "$50,001 - $100,000":
		c = 3
	case "$15,000 - $50,000", "$15,001 - $50,000":
		c = 2
	case "$1,001 - $15,000", "$1,000 - $15,000", "$1,001 -":
		c = 1
	default:
		return "🙈"
	}
	return strings.Repeat("💰", c)
}

func (dis Disclosure) ID() string {
	return strings.ToLower(
		strings.ReplaceAll(
			fmt.Sprintf(
				"%s_%s_%s_%s",
				dis.TransactionDate,
				dis.Type,
				dis.Ticker,
				dis.Representative,
			),
			" ",
			"_",
		),
	)
}

// TransactionOn gives the time that the trade was done
func (dis Disclosure) TransactionOn() time.Time {
	td, _ := time.Parse("01/02/2006", dis.TransactionDate)
	return td
}

// DisclosedOn gives the time that the trade was disclosed
func (dis Disclosure) DisclosedOn() time.Time {
	t, err := time.Parse("01/02/2006", dis.DisclosureDate)
	if err != nil {
		return time.Time{}
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func (dis Disclosure) DaysAgo() string {
	return fmt.Sprintf("%d", int((time.Now().Unix()-dis.TransactionOn().Unix())/86400))
}

func (dis Disclosure) TypeEmoji() string {
	var adj string
	switch strings.ToLower(dis.Type) {
	case "exchange":
		adj = "🔁"
	case "purchase":
		adj = "🤑"
	case "sale (full)":
		adj = "🤮"
	case "sale (partial)":
		adj = "🤢"
	default:
		adj = "🤷"
	}
	return adj
}

func (dis Disclosure) String() string {
	adj := dis.TypeEmoji()
	moneybags := dis.AmountEmojis()

	l1 := fmt.Sprintf("%s %s `%s` %s", dis.Representative, adj, dis.Ticker, moneybags)
	l2 := fmt.Sprintf("`%s` %s days ago (%s) totalling between %s", dis.AssetDescription, dis.DaysAgo(), dis.TransactionOn().Format("2006-01-02"), dis.Amount)
	s := fmt.Sprintf("%s\n%s", l1, l2)

	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	s = strings.ReplaceAll(s, "-", `\-`)

	s = bluemonday.StrictPolicy().Sanitize(s)

	return s
}

func GetDisclosuresFromFile(fn string) func() ([]Disclosure, error) {
	return func() ([]Disclosure, error) {
		var v []Disclosure
		data, err := ioutil.ReadFile(fn)
		if err != nil {
			return v, err
		}

		err = json.Unmarshal(data, &v)
		return v, err
	}
}

func DownloadDisclosuresFromS3(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		err = errors.New("unexpected status code " + res.Status)
		return nil, err
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func GetDisclosuresFromS3() (dd []Disclosure, err error) {
	data, err := DownloadDisclosuresFromS3(DisclosuresURLHouse)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &dd)
	if err != nil {
		return
	}

	data, err = DownloadDisclosuresFromS3(DisclosuresURLSenate)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &dd)
	return
}

// FromDate will return disclosures from an exact date
func FromDate(dd []Disclosure, date string) (do []Disclosure) {
	for _, d := range dd {
		if d.DisclosureDate == date {
			do = append(do, d)
		}
	}
	return
}

// Disclosures represents a collection of Disclosure objects
type Disclosures []Disclosure

// After will only return disclousres from the list after the given time
func (dd Disclosures) After(date Date) Disclosures {
	var out Disclosures

	for _, d := range dd {
		t := d.DisclosedOn()
		if t.After(date.Time()) {
			out = append(out, d)
		}
	}
	return out
}
