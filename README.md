# Tobot

# WIP - Work In Progress

Tob.lt bot, written in Go, inspired by Ansible modules and Telegraf plugin designs.

Features:
 * Intended for 24/7 uptime. It can be manually paused & resumed using Telegram bot.
 * Modular & customizable routines (examples [here](https://github.com/erkexzcx/tobot/tree/master/activities)).
 * Level-up multiple skills at the same time (full list of them [here](https://github.com/erkexzcx/tobot/tree/master/module)).
 * Automatically solves anti-bot checks (really, you won't even notice them)...
 * Automatically recover from network or `NUORODAS REIKIA SPAUSTI TIK VIENĄ KARTĄ!`-like errors.
 * Receive new PMs and reply back via Telegram bot.
 * Can be configured to automatically stop & resume at given customizable & randomized intervals/durations.

# Usage

1. [Install Golang](https://golang.org/doc/install) and dependencies. For Debian/Ubuntu:
```bash
apt install tesseract-ocr tesseract-ocr-lit libtesseract-dev gcc g++
```

Most of the "trusted" distros ships a very old Golang version in their official repositories, which might not work at all. Make sure to remove any existing Golang installations and install the latest version using [official upstream guide](https://golang.org/doc/install) for your operating system.

2. Build binary
```
# Build binary
go build -ldflags="-s -w" -o tobot ./cmd/tobot/main.go

# Verify if it's working
./tobot -help
```

3. Create configuration file. Simply copy `config.example.yml` to a new file `config.yml` and edit accordingly.

Telegram bot will notify you about important events (e.g. bot started or you got banned).

Telegram bot will also send you all received new PMs from the players. Reply to the player by simply **replying** to the same Telegram **bot's message**. Note that tob.lt bot **WILL STOP indefinitely** until you reply to the player. If you don't want to reply to the player, then reply to Telegram bot's message with text `/ignore`. Also note that Telegram bot will not send any message to the player that starts with `/`, so it's OK to make a TYPO mistake.

4. Create new directory, similar to existing one `activities` (use this dir as an example). Each file represents different activity, format must be `*.yml` and such files will be executed in alphabetical filename order (hence that's the meaning of `10_` in filenames). Once all activities are finished, bot will start from the top again. :)

Full list of modules: https://github.com/erkexzcx/tobot/tree/master/module

Non-module specific fields:
```
_module - (required) name of the module
_count - (optional) how many times perform the module action. 
```

All other fields are listed in README.md file within each module's directory.

5. Run program
```
./tobot
./tobot -config /path/to/config
./tobot -config /path/to/config -activities /path/to/activities_dir
./tobot -help
```
