# stonkcritter

A telegram bot that notifies you when a US congress critter makes a trade on the stock market at https://t.me/stonkcritter or with personal subscriptions via [@stonkcritter_bot](https://t.me/stonkcritter_bot)

Please note this does not index or tell you who has what stocks, only what stock trades have happened when they are disclosed.  For that, visit the terrific sites from which this bot gets it's updates:

* https://housestockwatcher.com
* https://senatestockwatcher.com

If you have any questions about the data this bot puts out and what it means, you'll probably want to check those sites.  For issues about the bot itself, create an issue.

## Level1Techs Devember 2021

[![level1techslogo](https://level1techs.com/sites/all/themes/l1/img/black-logo.png)](https://level1techs.com/)

This was done as part of the [Level1Techs Devember competition](https://forum.level1techs.com/t/official-devember-2021-welcome/177940).  Check them out at https://level1techs.com/ or https://www.youtube.com/c/level1techs

This is running on a Linode server using the Level1Techs coupon: https://linode.com/level1techs

[![image](https://user-images.githubusercontent.com/4642414/145663935-ca14c03f-c80f-4eaf-9dd4-141049720076.png)](https://linode.com/level1techs
)

## Features

- [x] broadcast messages to a Telegram Channel whenever a congress critter makes a trade
- [x] subscribe to specific stock tickers to see which congress critters are trading that stock
- [x] subscribe to a specific congress critter to see what trades they make
- [x] save the cursor to stop sending the same messages if the service restarts
- [x] get the direct message chat working
- [ ] follow multiple tickers with a single `/follow` message
- [ ] allow following specific asset types (e.g. stock options, crypto or futures)
- [x] allow sending to a NATS server
- [ ] allow sending to authenticated NATS server
- [x] allow serving disclosures via websockets
- [x] allow sending to a MQTT server
- [x] allow sending to an authenticated MQTT server
- [x] allow pushing to a remote webhook

## How to use

### Broadcast Channel

You can use the bot passively, just by joining the following channel, which will message about every single stock trade a congress critter does.

[![image](https://user-images.githubusercontent.com/4642414/145662011-826cf4a2-457e-4d4b-b806-7897b00991a5.png)](https://t.me/stonkcritter)

### Personalized subscriptions

You can use the bot actively, by messaging the bot directly, it's name is [@stonkcritter_bot](https://t.me/stonkcritter_bot).  Simply write `/help` to the bot and it will tell you what you can do, but the main
things to do are to follow congress critters or stock tickers (by prefixing a `$` to the symbol).  Try the following:

    /follow $MSFT
    /follow $TSLA
    /follow Nancy Pelosi

You can write `/list` at any time to see who you're following, with links to unfollow.

### Messages

![image](https://user-images.githubusercontent.com/4642414/145661696-f2f222b4-5ece-4107-a6c3-6251056366c6.png)

The messages show the congress critters name, what the trade was (purchase/sale etc as an emoji), the ticker and how much money was involved (as an emoji).  It also shows the date the transaction occured and the money figure range.  In most cases, the disclosure date will be the same day that you recieve the message in Telegram.

### Trade Type

These are the types that come out of the disclosure data:

| Sale Type | Emoji|
|---|---|
|exchange|🔁|
|purchase|🤑|
|sale (full)|🤮|
|sale (partial)|🤢|
|Unknown|🤷|

### Trade values

These are the trade values that come out of the disclosure data:

| Trade value | Emojis |
|---|---|
|$50,000,000 +|💰💰💰💰💰💰💰💰💰|
|$5,000,001 - $25,000,000|💰💰💰💰💰💰💰💰|
|$1,000,001 - $5,000,000|💰💰💰💰💰💰💰|
|$1,000,000 +|💰💰💰💰💰💰💰|
|$500,001 - $1,000,000|💰💰💰💰💰💰|
|$250,001 - $500,000|💰💰💰💰💰|
|$100,001 - $250,000|💰💰💰💰|
|$50,001 - $100,000|💰💰💰|
|$15,000 - $50,000|💰💰|
|$15,001 - $50,000|💰💰|
|$1,001 - $15,000|💰|
|$1,000 - $15,000|💰|
|$1,001 -|💰|
|Unknown|🙈|

## How to run it

Probably stick to the officially hosted bot, unless this repo becomes stale/unresponsive, in case we cause too much traffic for the websites
that host the source data.

    stonkcritter -h                                        # shows all the available options
    date --date="2022-01-01" "+%s" > ./stonkcritter.cursor # set the cursor to the start of 2022
    stonkcritter -download > transactions.json             # download all disclosures
    stonkcritter -f transactions.json -1                   # print all trade disclosures from the start of 2022 (and update the cursor)
    export BOT_TOKEN=yourtoken BOT_CHANNEL=yourchannelID
    stonkcritter -chat                                     # run the bot with the telegram interface

The cursor stores the current date that stonkcritter knows about disclosures up to.  By default it's kept in a file called `./stonkcritter.cursor`
and is simply a unix epoch timestamp.  If the file doesn't exist it will be automatically created with todays date.  You can specify the cursor file
 using `-c /path/to/your.cursor` if you don't want to use the default one.

Trades will be output to the terminal by default, if you don't want that use `-q` to shut it up.

The telegram bot direct message chat will not be activated unless the `-chat` flag and `BOT_TOKEN` environment variable is specified. If `BOT_CHANNEL` is not specified in the environment variables then the broadcasting is disabled.

The `-d ./brain` flag tells the bot where to store its database, or brain.  This is where things like congress critters and user subscriptions
are kept.

There is an informational API available using the `-api` flag which runs on `localhost:8090`.  The following endpoints are available:

* `GET /` - stats and the current cursor
* `GET /critters` - get a list of all the known congress critters
* `GET /subs` - show a list of all the subscriptions
* `PUT /watcher/check` - check the disclosure source immediately

### Disclosure sources

By default it will read from S3, unless you specify the `-f /path/to/disclosures.json` flag.  You can download the transactions like so:

    stonkcritter -download > disclosures.json

That makes it easier to test without hammering the stock watchers S3 bucket.

### Other sinks

You can also use the following "sinks" to send messages to by specifying their flags:

* NATS: `-n nats://localhost:4222/the.subject.to.publish.to`
* MQTT: `-m localhost:1883/the/topic/to/publish/to` and the environment variable `MQTT_CREDS=user:pass` if authenticated
* Websockets: `-w localhost:8080/ws/trades`

NATS and MQTT are clients and require a running broker to connect to, but the websockets one will start a server on the given
address/port and accept WS connections on the given path.

## How to run it after installing the debian package

### Config

You will need to set the token and channel ID via `BOT_TOKEN` and `BOT_CHANNEL` environment variables.  In the installed environment
this is found in `/etc/default/stonkcritter`:

    BOT_TOKEN=<bot_token>
    BOT_CHANNEL=<channel_id>

The default channel is included in the package but obviously only the bot with the correct token can post to that.

### Start the bot

Start and enable the bot on boot:

    systemctl enable stonkcritter
    systemctl start stonkcritter
