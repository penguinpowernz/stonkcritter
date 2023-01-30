package main

import (
	"time"

	"github.com/penguinpowernz/stonkcritter/models"
)

type critterMeta struct {
	Name                        string         `json:"name"`
	User                        string         `json:"user"`
	AssociatedTickers           []string       `json:"associated_tickers"`
	AssociatedAssets            []string       `json:"associated_assets"`
	PTRTrades                   int            `json:"ptr_trades"`
	ProperTrades                int            `json:"proper_trades"`
	TotalTrades                 int            `json:"total_trades"`
	AvgDisclosureDays           int            `json:"avg_disclosure_days"`
	AvgDisclosureDaysLast30Days int            `json:"avg_disclosure_days_last30_days"`
	AvgDisclosureDaysByMonths   map[string]int `json:"avg_disclosure_days_by_months"`
	TradesLast30Days            int            `json:"trades_last30_days"`
	TradesPerMonth              map[string]int `json:"trades_per_month"`
	Buys                        map[string]int `json:"buys"`
	Sells                       map[string]int `json:"sells"`
	PartialSells                map[string]int `json:"partial_sells"`
	Exchange                    map[string]int `json:"exchange"`

	avgDisclosureDaysLast30Days []int
	avgDisclosureDaysByMonths   map[string][]int
	avgDisclosureDays           []int
}

func mkCritterMeta(dd models.Disclosures) map[string]critterMeta {
	metas := map[string]critterMeta{}

	daysAgo30 := time.Now().Add(-30 * (time.Hour * 24))
	dd.Each(func(d models.Disclosure) {
		uname := parameterize(d.CritterName())
		days2disclose := int((d.DisclosedOn().Unix() - d.TransactionOn().Unix()) / 86400)

		var title string
		if d.Representative != "" {
			title = "Rep."
		}

		if d.Senator != "" {
			title = "Sen."

		}

		meta, found := metas[uname]
		if !found {
			meta = critterMeta{Name: title + " " + d.CritterName(), User: uname}
		}

		if meta.TradesPerMonth == nil {
			meta.TradesPerMonth = map[string]int{}
		}

		if meta.AvgDisclosureDaysByMonths == nil {
			meta.AvgDisclosureDaysByMonths = map[string]int{}
		}

		if meta.Buys == nil {
			meta.Buys = map[string]int{}
		}

		if meta.Sells == nil {
			meta.Sells = map[string]int{}
		}

		if meta.PartialSells == nil {
			meta.PartialSells = map[string]int{}
		}

		if meta.Exchange == nil {
			meta.Exchange = map[string]int{}
		}

		if meta.avgDisclosureDaysByMonths == nil {
			meta.avgDisclosureDaysByMonths = map[string][]int{}
		}

		meta.avgDisclosureDays = append(meta.avgDisclosureDays, days2disclose)

		meta.TotalTrades++
		meta.AssociatedTickers = append(meta.AssociatedTickers, d.TickerString())
		meta.AssociatedAssets = append(meta.AssociatedAssets, d.AssetTypeString())
		if d.IsPDFDisclosedFiling() {
			meta.PTRTrades++
		} else {
			meta.ProperTrades++
		}

		if d.TransactionOn().After(daysAgo30) {
			meta.TradesLast30Days++
			meta.avgDisclosureDaysLast30Days = append(meta.avgDisclosureDaysLast30Days, days2disclose)
		}

		mnth := d.TransactionOn().Format("200601")
		meta.TradesPerMonth[mnth]++
		meta.avgDisclosureDaysByMonths[mnth] = append(meta.avgDisclosureDaysByMonths[mnth], days2disclose)

		if list, found := meta.avgDisclosureDaysByMonths[mnth]; found {
			list = append(list, days2disclose)
			meta.avgDisclosureDaysByMonths[mnth] = list
		} else {
			meta.avgDisclosureDaysByMonths[mnth] = []int{days2disclose}
		}

		switch d.TradeType() {
		case "purchase":
			meta.Buys[mnth]++
		case "sale_full":
			meta.Sells[mnth]++
		case "sale_partial":
			meta.PartialSells[mnth]++
		case "exchange":
			meta.Exchange[mnth]++
		}

		metas[uname] = meta
	})

	for uname, meta := range metas {
		meta.AvgDisclosureDaysLast30Days = avg(meta.avgDisclosureDaysLast30Days)
		meta.AvgDisclosureDays = avg(meta.avgDisclosureDays)
		for m, num := range meta.avgDisclosureDaysByMonths {
			meta.AvgDisclosureDaysByMonths[m] = avg(num)
		}

		meta.AssociatedAssets = unique(meta.AssociatedAssets)
		meta.AssociatedTickers = unique(meta.AssociatedTickers)

		metas[uname] = meta
	}

	return metas
}
