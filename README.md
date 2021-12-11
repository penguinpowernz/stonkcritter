# stonkcritter

A telegram bot that notifies you when a US congress critter makes a trade on the stock market.

* The official channel is here: https://t.me/stonkcritter

Note this does not index or tell you who has what stocks.  For that, visit the terrific sites from which this bot gets it's updates:

* https://housestockwatcher.com
* https://senatestockwatcher.com

If you have any questions about the data this bot puts out and what it means, you'll probably want to check those sites.  For issues about the bot itself, create an issue.

## Messages

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

## Usage

### Spam stop

You'll need to set the cursor manually to start with, this will stop it dumping all disclosures ever:

    $ stonkcritter -x 2021-10-30
    2021/12/11 00:46:46 parsed cursor time as 2021-10-30 00:00:00 +0000 UTC
    2021/12/11 00:46:46 updated cursor to 2021-10-30 (2021-10-30 00:00:00 +0000 UTC)

Or in the installed environment:

    sudo -iu stonkcritter stonkcritter -x 2021-10-30 -d /home/stonkcritter/data

### Config

You will need to set the token and channel ID via `BOT_TOKEN` and `BOT_CHANNEL`.  In the installed environment
this is found in `/etc/default/stonkcritter`:

    BOT_TOKEN=<bot_token>
    BOT_CHANNEL=<channel_id>

The default channel is included in the package but obviously only the bot with the correct token can post to that.

### Testing

You can dry run by not specifying the `-chat` arg:

    stonkcritter

That will log all disclosures to the terminal instead of broadcasting.

### Start the bot

Start and enable the bot on boot:

    systemctl enable stonkcritter
    systemctl start stonkcritter

You can also run it from command line:

    BOT_TOKEN=xxx BOT_CHANNEL=xxx stonkcritter -chat

Omitting BOT_CHANNEL will cause the bot not to broadcast disclosures to channel.

### API

A local API is running on port 8090 with the following endpoints:

* `GET  http://127.0.0.1:8090/reps` - all known reps
* `PUT  http://127.0.0.1:8090/disclosures?cursor=xx` - push a new disclosures file in and set the cursor
* `POST http://127.0.0.1:8090/pull_from_s3` - download disclosures from S3

You can run the following commands to call these on the running API on the same system:

    $ stockcritter -pull
    $ stockcritter -loadfile <filename.json> [-x <2021-10-30>]

### Chat interface

WORK IN PROGRESS

There are a few commands:

* `/follow` to follow a ticker to congress critter
* `/list` to show you what you are following
* `/unfollow` to stop following something
* `/findrep` to search for a specific representative
