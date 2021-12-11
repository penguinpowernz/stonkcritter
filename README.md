# politstonk

A telegram bot that notifies you when a US congress critter makes a trade on the stock market

## Usage

### Spam stop

You'll need to set the cursor manually to start with, this will stop it dumping all disclosures ever:

    stonkcritter -x 2021-10-30

Or in the installed environment:

    sudo -iu stonkcritter stonkcritter -x 2021-10-30 -d /home/stonkcritter/data

### Config

You will need to set the token and channel ID via `BOT_TOKEN` and `BOT_CHANNEL`.  In the installed environment
this is found in `/etc/default/stonkcritter`:

    BOT_TOKEN=<bot_token>
    BOT_CHANNEL=<channel_id>

The default channel is included in the package but obviously only the bot with the correct token can post to that.

### Start the bot

Start and enable the bot on boot:

    systemctl enable stonkcritter
    systemctl start stonkcritter

You can also run it from command line:

    BOT_TOKEN=xxx BOT_CHANNEL=xxx stonkcritter -chat

Omitting BOT_CHANNEL will cause the bot not to broadcast disclosures to channel.

## Bot Usage

### API

A local API is running on port 8090 with the following endpoints:

* `http://127.0.0.1:8090/reps` - all know reps
* `http://127.0.0.1:8090/disclosures` - all know disclosures

### Chat interface

WORK IN PROGRESS

There are a few commands:

* `/follow` to follow a ticker to congress critter
* `/list` to show you what you are following
* `/unfollow` to stop following something
* `/findrep` to search for a specific representative
