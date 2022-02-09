package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/penguinpowernz/stonkcritter/api"
	"github.com/penguinpowernz/stonkcritter/bot"
	"github.com/penguinpowernz/stonkcritter/models"
	SINKS "github.com/penguinpowernz/stonkcritter/sinks"
	"github.com/penguinpowernz/stonkcritter/source"
	"github.com/penguinpowernz/stonkcritter/watcher"
)

var (
	dataDir    = "./data"
	fileSource = ""
	cursorFile = "./stonkcritter.cursor"
	wsURL      = os.Getenv("WS_URL")
	natsURL    = os.Getenv("NATS_URL")
	mqttURL    = os.Getenv("MQTT_URL")
	webhookURL = os.Getenv("WEBHOOK_URL")

	runAPI, runChat, quiet, downloadDump, runOnce bool
)

func main() {
	flag.StringVar(&cursorFile, "c", cursorFile, "the file the current cursor is saved to")
	flag.StringVar(&dataDir, "d", dataDir, "the directory to save bot brain data in")
	flag.StringVar(&fileSource, "f", fileSource, "read from the given source file instead of S3")
	flag.StringVar(&wsURL, "w", wsURL, "the websockets URL and path (e.g. 127.0.0.1:8080/ws)")
	flag.StringVar(&mqttURL, "m", mqttURL, "the MQTT URL and path (e.g. 127.0.0.1:1833/stonk/critter/trades)")
	flag.StringVar(&natsURL, "n", natsURL, "the NATS URL and subject (e.g. nats://127.0.0.1:4222/stonk.critter.trades)")
	flag.StringVar(&webhookURL, "k", webhookURL, "the webhook URL to post to (e.g. http://example.com/stonks)")
	flag.BoolVar(&runChat, "chat", false, "enable Telegram communication")
	flag.BoolVar(&runAPI, "api", false, "enable informational API")
	flag.BoolVar(&quiet, "q", false, "don't log disclosure messages to terminal")
	flag.BoolVar(&runOnce, "1", false, "only run a single check")
	flag.BoolVar(&downloadDump, "download", false, "download the disclosures and dump them to STDOUT")
	flag.Parse()

	if downloadDump {
		downloadAndDump()
		os.Exit(0)
	}

	var opts []watcher.Option
	if fileSource != "" {
		opts = append(opts, watcher.FromFile(fileSource))
		log.Println("using file disclosure source", fileSource)
	} else {
		opts = append(opts, watcher.FromS3())
		log.Println("using S3 disclosure source")
	}

	if cursorFile != "" {
		setCursorToTodayIfNotExist(cursorFile)
		opts = append(opts, watcher.DiskCursor(cursorFile, true))
		log.Println("using cursor file", cursorFile)
	}

	w, err := watcher.NewWatcher(opts...)
	if err != nil {
		panic(err)
	}

	var sinks []SINKS.Sink

	if !quiet {
		sinks = append(sinks, SINKS.Writer(os.Stdout))
		log.Println("added sink: stdout")
	}

	createNATSSink(&sinks)
	createMQTTSink(&sinks)
	createWebsocketSink(&sinks)
	createWebhookSink(&sinks)
	createBroadcastSink(&sinks)
	createBotSink(&sinks, w)

	log.Println("starting the disclosure watcher")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// we don't want the rate limiting of some things to affect other things so start a pool
	pool := NewPool(40)
	w.Start(ctx)
	for w.Next() {
		d := w.Disclosure()

		for i, sink := range sinks {
			log.Printf("sending %s to sink %d", d.ID(), i)

			// must call these inside the function to prevent variable scoping across loops
			func(sink SINKS.Sink, d models.Disclosure) {
				pool.Run(func() { sink(d) })
			}(sink, d)
		}

		if runOnce && w.Checks() > 0 && w.Inflight() == 0 {
			pool.wg.Wait()
			break
		}
	}
}

func startAPI(w *watcher.Watcher, b *bot.Brain, bt *bot.Bot) {
	server := api.NewServer(struct {
		*watcher.Watcher
		*bot.Brain
		*bot.Bot
	}{w, b, bt})
	r := gin.Default()
	server.SetupRoutes(r)
	r.Run(":8090")
}

func setCursorToTodayIfNotExist(fn string) {
	_, err := os.Stat(fn)
	if os.IsNotExist(err) {
		ioutil.WriteFile(fn, []byte(strconv.Itoa(int(time.Now().Unix()))), 0644)
	}
}

func downloadAndDump() {
	dd, err := source.GetDisclosuresFromS3()
	if err != nil {
		panic(err)
	}
	data, err := json.Marshal(dd)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(data)
}

func createBotSink(sinks *[]SINKS.Sink, w *watcher.Watcher) {
	if os.Getenv("BOT_TOKEN") == "" || !runChat {
		return
	}

	brain, err := bot.NewBrain(dataDir)
	if err != nil {
		panic(err)
	}

	bot, err := bot.NewBot(brain, os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	// update the critters in the bots brain after every check
	go func() {
		for {
			w.WaitForCheck()
			brain.StoreCritters(w.Critters())
		}
	}()

	*sinks = append(*sinks, SINKS.TelegramBot(bot))
	log.Println("added sink: telegram bot")

	if runAPI {
		log.Println("starting informational API")
		go startAPI(w, brain, bot)
	}
}

func createMQTTSink(sinks *[]SINKS.Sink) {
	if mqttURL == "" {
		return
	}

	topic := "stonk/critter/trades"
	bits := strings.Split(mqttURL, "/")
	url := bits[0]
	if len(bits) > 1 {
		topic = strings.Join(bits[1:], "/")
	}

	if !strings.HasPrefix(topic, "/") {
		topic = "/" + topic
	}

	sink, err := SINKS.MQTT(url, os.Getenv("MQTT_CREDS"), topic)
	if err != nil {
		log.Println("ERROR: failed to create MQTT sink:", err)
		return
	}

	log.Printf("creating MQTT sink connected to %s, publishing on topic %s", url, topic)
	*sinks = append(*sinks, sink)
}

func createNATSSink(sinks *[]SINKS.Sink) {
	if natsURL == "" {
		return
	}

	subj := "stonk.critter.trades"
	bits := strings.Split(natsURL, "/")
	url := bits[0]
	if len(bits) > 1 {
		subj = bits[1]
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return
	}

	log.Printf("creating NATS sink connected to %s, publishing on topic %s", url, subj)
	*sinks = append(*sinks, SINKS.NATS(nc, subj))
}

func createWebhookSink(sinks *[]SINKS.Sink) {
	if webhookURL == "" {
		return
	}

	log.Printf("creating webhook sink connected to %s", webhookURL)
	*sinks = append(*sinks, SINKS.Webhook(webhookURL))
}

func createWebsocketSink(sinks *[]SINKS.Sink) {
	if wsURL == "" {
		return
	}

	sink, err := SINKS.Websockets(wsURL)
	if err != nil {
		return
	}

	log.Printf("creating WS sink running on %s", wsURL)
	*sinks = append(*sinks, sink)
}

func createBroadcastSink(sinks *[]SINKS.Sink) {
	if os.Getenv("BOT_TOKEN") == "" || os.Getenv("BOT_CHANNEL") == "" {
		return
	}

	broadcast, err := SINKS.TelegramChannel(os.Getenv("BOT_TOKEN"), os.Getenv("BOT_CHANNEL"))
	if err != nil {
		log.Println("ERROR: failed to parse BOT_CHANNEL:", err)
		return
	}

	*sinks = append(*sinks, broadcast)
	log.Println("added sink: telegram broadcast to channel", os.Getenv("BOT_CHANNEL"))
}

type Pool struct {
	tokens chan struct{}
	wg     *sync.WaitGroup
}

func NewPool(c int) *Pool {
	return &Pool{make(chan struct{}, c), new(sync.WaitGroup)}
}

func (p *Pool) Run(op func()) {
	p.tokens <- struct{}{}
	p.wg.Add(1)
	go func() {
		defer func() { <-p.tokens }()
		defer p.wg.Done()
		op()
	}()
}
