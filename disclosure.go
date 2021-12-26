package politstonk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

var DisclosuresURLHouse = "https://house-stock-watcher-data.s3-us-west-2.amazonaws.com/data/all_transactions.json"
var DisclosuresURLSenate = "https://senate-stock-watcher-data.s3-us-west-2.amazonaws.com/aggregate/all_transactions.json"

//   COUNT  TYPES OF OWNER
//       1  "FL HSG Fin Corp Homeowner MTS Rev 2.15% Due 07/01/2029";
//       1  "Florida HSg Fin Corp Homeowner 2.15% Due 07/01/29";
//       1  "Florida HSG Fin Corp Homeowner MTG 2.15%";
//       1  "Florida HSg Fin Corp Homeowner MTg 2.15% Due 07/01/29";
//       1  "Florida HSG Fin Corp Homeowner MTG Rev 2.15% Due 07/01/2029";
//       1  "Florida HSG Fin Corp Homeowner MTG Rev 2.15% Due 7/1/2029";
//     176  "Child";
//     375  "dependent";
//     465  "N/A";
//    1315  "--";
//    1353  "Self";
//    2633  "self";
//    3374  "Spouse";
//    3592  "Joint";
//    3767  "joint";

//   COUNT  TYPES OF ASSET
//       3  "Cryptocurrency";
//      20  "Commodities/Futures Contract";
//      97  "Non-Public Stock";
//     203  "Stock Option";
//     237  "Corporate Bond";
//     361  "Other Securities";
//     386  "Municipal Security";
//     465  "PDF Disclosed Filing";
//    6522  "Stock";

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

// {
// 	"transaction_date": "11/01/2021",
// 	"owner": "Self",
// 	"ticker": "PYPL",
// 	"asset_description": "PayPal Holdings, Inc. - Common Stock",
// 	"asset_type": "Stock",
// 	"type": "Sale (Full)",
// 	"amount": "$250,001 - $500,000",
// 	"comment": "--",
// 	"senator": "John W Hickenlooper",
// 	"ptr_link": "https://efdsearch.senate.gov/search/view/ptr/3ca89f70-6cd2-4b06-a15d-44fd54fc58fa/",
// 	"disclosure_date": "12/10/2021"
// }

type Disclosure struct {
	DisclosureYear     int    `json:"disclosure_year,omitempty"`
	DisclosureDate     string `json:"disclosure_date,omitempty"`
	TransactionDate    string `json:"transaction_date,omitempty"`
	Owner              string `json:"owner,omitempty"`
	Ticker             string `json:"ticker,omitempty"`
	AssetDescription   string `json:"asset_description,omitempty"`
	Type               string `json:"type,omitempty"`
	AssetType          string `json:"asset_type,omitempty"`
	Amount             string `json:"amount,omitempty"`
	Representative     string `json:"representative,omitempty"`
	Senator            string `json:"senator,omitempty"`
	District           string `json:"district,omitempty"`
	PtrLink            string `json:"ptr_link,omitempty"`
	CapGainsOver200Usd bool   `json:"cap_gains_over_200_usd,omitempty"`
}

func (dis Disclosure) CritterTopic() string {
	return dis.CritterName()
}

func (dis Disclosure) TickerTopic() string {
	return "$" + dis.Ticker
}

func (dis Disclosure) AssetTypeTopic() string {
	switch dis.AssetType {
	case "Cryptocurrency":
		return "#crypto"
	case "Commodities/Futures Contract":
		return "#comfuture"
	case "Non-Public Stock":
		return "#nopstock"
	case "Stock Option":
		return "#opt"
	case "Corporate Bond":
		return "#corpbond"
	case "Other Securities":
		return "#osec"
	case "Municipal Security":
		return "#msec"
	case "PDF Disclosed Filing":
		return "#pdf"
	default:
		return "#stonk"
	}
}

func (dis Disclosure) OwnerString() string {
	switch strings.ToLower(dis.Owner) {
	case "self":
		return ""
	case "joint":
		return ", joint owned"
	case "dependent":
		return ", owned by their dependent"
	case "spouse":
		return ", owned by their spouse"
	case "child":
		return ", owned by their child"
	case "--", "n/a":
		return ", it is known who exactly owns this"
	default:
		return ", owned by" + dis.Owner
	}
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
	return normalizeDate(dis.TransactionDate)
}

// DisclosedOn gives the time that the trade was disclosed
func (dis Disclosure) DisclosedOn() time.Time {
	t := normalizeDate(dis.DisclosureDate)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func (dis Disclosure) DaysAgo() string {
	return fmt.Sprintf("%d", int((time.Now().Unix()-dis.TransactionOn().Unix())/86400))
}

func (dis Disclosure) IsPDFDisclosedFiling() bool {
	return dis.AssetType == "PDF Disclosed Filing"
}

func (dis Disclosure) AssetTypeString() string {
	if dis.AssetType != "Stock" {
		return " (" + dis.AssetType + ")"
	}
	return ""
}

func (dis Disclosure) TickerString() string {
	switch dis.Ticker {
	case "N/A", "--":
		return "??"
	default:
		return dis.Ticker
	}
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

func (dis Disclosure) CritterName() string {
	critter := dis.Representative
	if critter == "" {
		critter = dis.Senator
	}
	return critter
}

func (dis Disclosure) DodgeyFilingString() string {
	return fmt.Sprintf(
		"%s did a dodgey... they've filed a PDF which is not parsed, so you'll have to check %s for the details",
		dis.CritterName(),
		dis.PtrLink,
	)
}

func (dis Disclosure) NormalString() string {
	adj := dis.TypeEmoji()
	moneybags := dis.AmountEmojis()

	l1 := fmt.Sprintf("%s %s `%s` %s", dis.CritterName(), adj, dis.TickerString(), moneybags)
	l2 := fmt.Sprintf("`%s` %s days ago (%s) totalling between %s", dis.AssetDescription, dis.DaysAgo(), dis.TransactionOn().Format("2006-01-02"), dis.Amount)
	s := fmt.Sprintf("%s\n%s", l1, l2)

	switch dis.AssetType {
	case "Stock", "PDF Disclosed Filing":
	default:
		l3 := fmt.Sprintf("\n*%s%s*", dis.AssetType, dis.OwnerString())
		s += l3
	}

	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	s = strings.ReplaceAll(s, "-", `\-`)

	s = bluemonday.StrictPolicy().Sanitize(s)

	return s
}

func (dis Disclosure) String() string {
	if dis.IsPDFDisclosedFiling() {
		return dis.DodgeyFilingString()
	}

	return dis.NormalString()
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
	log.Println("got data from", url)
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
	var _dd []Disclosure
	err = json.Unmarshal(data, &_dd)
	dd = append(dd, _dd...)
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
