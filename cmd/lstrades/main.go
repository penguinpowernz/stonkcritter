package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/penguinpowernz/stonkcritter/models"
	"github.com/penguinpowernz/stonkcritter/renderers"
	"github.com/penguinpowernz/stonkcritter/source"
)

var defaultFile = "stonkcritter.json"

func main() {

	var cursor, file string
	var download, mastodon, printCritters, printTickers, printReport, fmtJSON, fmtCSV bool
	flag.StringVar(&file, "f", defaultFile, "show trades from a specific time, use 't:' and 'd:' for trade and disclosure date, eg: (t:2021-09-29)")
	flag.BoolVar(&download, "d", false, "Download and save the latest trades to "+defaultFile)
	flag.StringVar(&cursor, "c", "", "show trades from a specific time, use 't:' and 'd:' for trade and disclosure date, eg: (t:2021-09-29)")
	flag.BoolVar(&mastodon, "m", false, "output in mastodon ruby console code")
	flag.BoolVar(&printCritters, "r", false, "output only the members names")
	flag.BoolVar(&printTickers, "t", false, "output only the ticker symbols")
	flag.BoolVar(&printReport, "w", false, "output ticker activity report")
	flag.BoolVar(&fmtJSON, "json", false, "output in JSON")
	flag.BoolVar(&fmtCSV, "csv", false, "output in CSV")
	flag.Parse()

	var dd models.Disclosures
	var err error

	var w io.Writer
	if defaultFile == "-" {
		w = os.Stdout
	}

	fileExist := func() bool { _, err := os.Stat(defaultFile); return !os.IsNotExist(err) }()
	if download || !fileExist {
		dd, err = source.GetDisclosuresFromS3()
		if err != nil {
			panic(err)
		}

		data, err := json.Marshal(dd)
		if err != nil {
			panic(err)
		}

		w, err = os.OpenFile(defaultFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
		if err != nil {
			panic(err)
		}

		if _, err := w.Write(data); err != nil {
			panic(err)
		}
		return
	}

	if dd == nil {
		dd, err = source.GetDisclosuresFromFile(defaultFile)()
		if err != nil {
			panic(err)
		}
	}

	bits := strings.Split(cursor, ":")

	fromField := "t"
	if len(bits) > 1 {
		fromField = bits[0]
		bits[0] = bits[1]
	}
	fromDate := bits[0]
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		panic(err)
	}

	render := render1()

	dd = models.Disclosures(dd).Filter(func(d models.Disclosure) bool {
		t := d.TransactionOn()
		if fromField == "d" {
			t = d.DisclosedOn()
		}
		return t.After(from)
	})

	critters := dd.Critters()
	if printCritters {
		for _, name := range critters {
			fmt.Println(name)
		}
		return
	}

	if mastodon {
		renderers.Mastodon(os.Stdout, dd)
		return
	}

	if printTickers {
		for _, name := range dd.Tickers() {
			fmt.Println(name)
		}
		return
	}

	if printReport {
		rep, typs := generateReport(dd)

		if fmtJSON {
			data, _ := json.MarshalIndent(rep, "", "  ")
			os.Stdout.Write(data)
			return
		}

		fmt.Printf("%10s", "name")
		for _, key := range typs {
			if fmtCSV {
				fmt.Printf(",%s", key)
				continue
			}

			fmt.Printf(" %10s", key)
		}
		fmt.Println()

		for name, counts := range rep {
			fmt.Printf("%10s", name)
			for _, key := range typs {
				if fmtCSV {
					fmt.Printf(",%d", counts[key])
					continue
				}

				fmt.Printf(" %10d", counts[key])
			}
			fmt.Println()
		}
		return
	}

	dd.Each(func(d models.Disclosure) {
		fmt.Println(render(d))
	})
}

func render1() func(d models.Disclosure) string {
	return func(d models.Disclosure) string {
		return fmt.Sprintf("%s %s %s days ago %s %s %s %s", d.CritterName(), d.TypeEmoji(), d.DaysAgo(), d.AmountEmojis(), d.TickerString(), d.Owner, d.OwnerString())
	}
}

func generateReport(dd models.Disclosures) (map[string]map[string]int, []string) {
	typs := []string{}
	uniq := func(n string) bool {
		for _, _n := range typs {
			if _n == n {
				return false
			}
		}
		return true
	}

	rep := map[string]map[string]int{}
	dd.Each(func(d models.Disclosure) {
		t, found := rep[d.Ticker]
		if !found {
			t = map[string]int{}
		}

		if uniq(d.TradeType()) {
			typs = append(typs, d.TradeType())
		}

		t[d.TradeType()]++
		rep[d.Ticker] = t
	})

	return rep, typs
}
