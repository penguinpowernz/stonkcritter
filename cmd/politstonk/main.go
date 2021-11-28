package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
	"github.com/penguinpowernz/politstonk"
	"github.com/timshannon/badgerhold/v4"
)

func main() {
	var logBC bool
	var setCursor, fileSource, dataDir string
	flag.StringVar(&dataDir, "d", "./data", "the directory to save bot brain data in")
	flag.StringVar(&fileSource, "f", "./all_transactions.json", "read from the given source file")
	flag.StringVar(&setCursor, "x", "", "set the current cursor, YYYY-MM-DD")
	flag.BoolVar(&logBC, "n", false, "only log broadcast messages, don't send to Telegram")
	flag.Parse()

	opts := badgerhold.DefaultOptions
	opts.Options = badger.DefaultOptions(dataDir)

	brain, err := badgerhold.Open(opts)
	if err != nil {
		panic(err)
	}

	// update the cursor so we don't have to care about storing
	// or broadcasting every single discloure ever
	if setCursor != "" {
		t, err := time.Parse("2006-01-02", setCursor)
		if err != nil {
			log.Fatalf("failed to parse the cursor: %s: %s", setCursor, err)
		}
		log.Printf("parsed cursor time as %s", t)
		d := politstonk.NewDate(t)
		if err := brain.Upsert("cursor", &d); err != nil {
			log.Fatalf("failed to save the cursor: %s: %s", setCursor, err)
		}
		log.Printf("updated cursor to %s (%s)", d.S, d.Time())
		os.Exit(0)
	}

	bot, err := politstonk.NewBot(
		brain,
		os.Getenv("BOT_TOKEN"),
		os.Getenv("BOT_CHANNEL"),
	)

	bot.LogOnly = logBC

	if err != nil {
		panic(err)
	}

	api := gin.Default()

	api.PUT("/disclosures", bot.HandleDisclosures)
	api.GET("/reps", bot.HandleListReps)
	go api.Run("localhost:8090")

	// use the
	var discloser func() ([]politstonk.Disclosure, error)
	discloser = politstonk.GetDisclosuresFromS3
	if fileSource != "" {
		discloser = politstonk.GetDisclosuresFromFile(fileSource)
	}

	t := time.NewTicker(time.Hour * 24)
	for {
		dd, err := discloser()
		if err != nil {
			panic(err)
		}

		bot.ConsumeDisclosures(dd)

		<-t.C
	}
}
