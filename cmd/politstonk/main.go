package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
	"github.com/penguinpowernz/politstonk"
	"github.com/timshannon/badgerhold/v4"
)

func main() {
	var runChat, downloadDump, pull bool
	var setCursor, fileSource, dataDir, loadFile string
	flag.StringVar(&dataDir, "d", "./data", "the directory to save bot brain data in")
	flag.StringVar(&fileSource, "f", "", "read from the given source file instead of S3")
	flag.StringVar(&setCursor, "x", "", "set the current cursor, YYYY-MM-DD")
	flag.BoolVar(&runChat, "chat", false, "enable Telegram communication")
	flag.BoolVar(&downloadDump, "download", false, "download the disclosures and dump them to STDOUT")
	flag.BoolVar(&pull, "pull", false, "trigger the HTTP API to pull the disclosures from S3")
	flag.StringVar(&loadFile, "loadfile", "", "load the disclosures in the given file via the running bots HTTP API (combine with -x to dump from cursor)")
	flag.Parse()

	log.SetOutput(os.Stderr)

	if downloadDump {
		downloadAndDump()
		os.Exit(0)
	}

	if loadFile != "" {
		loadFileAndCursor(loadFile, setCursor)
		os.Exit(0)
	}

	if pull {
		pullS3()
		os.Exit(0)
	}

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

	if err != nil {
		panic(err)
	}

	bot.LogOnly = !runChat
	api := gin.Default()

	api.PUT("/disclosures", bot.HandleDisclosures)
	api.GET("/reps", bot.HandleListReps)
	api.PUT("/cursor/:cursor", bot.HandleSetCursor)
	api.POST("/pull_from_s3", bot.HandlePullFromS3)
	go api.Run("localhost:8090")

	// use the
	var discloser func() ([]politstonk.Disclosure, error)
	discloser = politstonk.GetDisclosuresFromS3
	if fileSource != "" {
		discloser = politstonk.GetDisclosuresFromFile(fileSource)
	}

	t := time.NewTicker(time.Hour * 24)
	for {
		<-t.C
		dd, err := discloser()
		if err != nil {
			panic(err)
		}

		bot.ConsumeDisclosures(dd)
	}
}

func downloadAndDump() {
	dd, err := politstonk.GetDisclosuresFromS3()
	if err != nil {
		panic(err)
	}
	data, err := json.Marshal(dd)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(data)
}

func loadFileAndCursor(fn string, cursor string) {
	r, err := os.Open(fn)
	if err != nil {
		panic(err)
	}

	url := "http://localhost:8090/disclosures"
	if cursor != "" {
		url += "?cursor=" + cursor
	}

	req, err := http.NewRequest(http.MethodPut, url, r)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	if res.StatusCode == http.StatusNoContent {
		fmt.Println("OK")
		return
	}

	fmt.Println("Bad status returned:", res.StatusCode)
}

func pullS3() {
	res, err := http.Post("http://localhost:8090/pull_from_s3", "", nil)
	if err != nil {
		panic(err)
	}

	if res.StatusCode == http.StatusNoContent {
		fmt.Println("OK")
		return
	}
}
