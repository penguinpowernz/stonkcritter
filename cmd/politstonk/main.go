package main

import (
	"encoding/json"
	"flag"
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

		for _, d := range dd {
			bot.Broadcast(d.String())
			bot.HandleDisclosure(d)
			time.Sleep(time.Second)
		}
		<-t.C
	}
}

func readFile(fn string) ([]politstonk.Disclosure, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		panic(err)
	}

	var v []politstonk.Disclosure
	err = json.Unmarshal(data, &v)
	return v, err
}
