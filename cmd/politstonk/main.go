package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/penguinpowernz/politstonk"
)

func main() {
	var readFromS3 bool
	var sourceFile string
	flag.BoolVar(&readFromS3, "-s", false, "read from the S3 source file")
	flag.StringVar(&sourceFile, "-f", "", "read from the given source file")
	flag.Parse()

	var bcChannel int32
	bot, err := politstonk.NewBot(os.Getenv("BOT_TOKEN"), bcChannel)
	if err != nil {
		panic(err)
	}

	t := time.NewTicker(time.Hour * 24)
	for {
		var dd []politstonk.Disclosure
		var err error

		// we want to be able to read from a file or S3, so we don't hammer the
		// providers S3 egress costs during testing
		switch {
		case sourceFile != "":
			dd, err = readFile(sourceFile)
		case readFromS3:
			dd, err = politstonk.GetDisclosures()
		}

		if err != nil {
			panic(err)
		}

		bot.StoreReps(dd)

		// only get the disclosures from today
		dd = politstonk.FromDate(dd, time.Now().Format("02/01/2006"))

		fmt.Println("Found", len(dd), "disclosures from today")

		for _, d := range dd {
			bot.Broadcast(d.String())
			bot.HandleDisclosure(d)
			time.Sleep(time.Second)
		}
		<-t.C
	}
}

func readFile(fn string) ([]politstonk.Disclosure, error) {
	var v []politstonk.Disclosure
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return v, err
	}

	err = json.Unmarshal(data, &v)
	return v, err
}
