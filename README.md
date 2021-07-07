# Tobot

# WIP - Work In Progress

Tob.lt bot, written in Go, inspired by Ansible modules and Telegraf plugin designs.

Main features:
 * Intended for 24/7 uptime. It can be paused & resumed using Telegram bot.
 * Modular & customizable routines (see them [here](https://github.com/erkexzcx/tobot/tree/main/module)).
 * Level-up multiple skills at the same time.
 * Automatically solves anti-bot checks (really, you won't even notice them)...
 * Automatically recover from network or `NUORODAS REIKIA SPAUSTI TIK VIENĄ KARTĄ!`-like errors.
 * Receive new PMs and reply back via Telegram bot.
 * Can be configured to automatically stop & resume at given customizable & randomized intervals/durations.

# Usage

1. Install dependencies. For Debian/Ubuntu:
```bash
apt install tesseract-ocr tesseract-ocr-lit libtesseract-dev gcc g++
```

Most of the "trusted" distros ships a very old Golang version which might not be able to build binary. To avoid issues, remove any existing Golang installations and install the latest version using official upstream guide.

2. Build binary
```
# Build binary
go build -ldflags="-s -w" -o tobot ./cmd/tobot/main.go

# Verify if it's working
./tobot -help
```

3. Create configuration file. Simply copy `config.example.yml` to a new file `config.yml` and edit accordingly.

Telegram bot will:
 * Notify you about important events (e.g. you got banned);
 * Player sent you a PM. Reply to Telegram bot's message in order to reply the same text to the player. Note that bot **WILL STOP indefinitely** until you reply to the player. If you don't want to reply, just *to the Telegram **bot's message*** the text `/ignore`. Don't worry, any text that starts with `/` will not be sent to player, so it's okay to write `/igore`. :)

4. Create new directory, similar to existing one `activities` (use this dir as an example). Each file represents different activity, format must be `*.yml` and such files will be executed in alphabetical filename order (hence that's the meaning of `10_` in filenames). Once all activities are finished, bot will start from the top again. :)

Full list of modules: https://github.com/erkexzcx/tobot/tree/main/module

Mandatory fields for each task:
```
module - name of the module
```

Optional fields:
```
count - perform task this amount of times. Useful to limit "slayer" or cut just 1 wood for "kepimas" step.
```

5. Run program
```
./tobot
./tobot -config /path/to/config
./tobot -config /path/to/config -activities /path/to/activities_dir
./tobot -help
```
