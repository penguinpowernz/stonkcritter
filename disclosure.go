package politstonk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var DisclosuresURL = "https://house-stock-watcher-data.s3-us-west-2.amazonaws.com/data/all_transactions.json"

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

func (dis Disclosure) DaysAgo() string {
	td, _ := time.Parse("2006-01-02", dis.TransactionDate)
	return fmt.Sprintf("%d", int((time.Now().Unix()-td.Unix())/86400))
}

func (dis Disclosure) String() string {
	adj := "sold"
	if dis.Type == "purchase" {
		adj = "bought"
	}
	return fmt.Sprintf("%s %s (%s) %s on %s days ago (%s) totalling between %s", dis.Representative, adj, dis.Ticker, dis.AssetDescription, dis.DaysAgo(), dis.TransactionDate, dis.Amount)
}

func GetDisclosures() (dd []Disclosure, err error) {
	res, err := http.Get(DisclosuresURL)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = errors.New("unexpected status code " + res.Status)
		return
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(data, &dd)
	return
}

func FromDate(dd []Disclosure, date string) (do []Disclosure) {
	for _, d := range dd {
		if d.DisclosureDate == date {
			do = append(do, d)
		}
	}
	return
}
