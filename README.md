# Tobot

Tob.lt bot, written in Go, inspired by Ansible modules and Telegraf plugin designs.

Features:
 * Intended for 24/7 uptime (for autoamted stop/resume one should use `crontab` to start/stop this application).
 * Multi-user support.
 * Modular & customizable routines (examples [here](https://github.com/erkexzcx/tobot/tree/master/activities)).
 * Level-up multiple skills at the same time (full list of them [here](https://github.com/erkexzcx/tobot/tree/master/module)).
 * Automatically solves anti-bot checks (really, you won't even notice them)...
 * Automatically recover from network or `NUORODAS REIKIA SPAUSTI TIK VIENĄ KARTĄ!`-like errors.
 * Receive new PMs and reply back via Telegram bot.
 * Maximum clicking performance, uses your provided RTT duration to ensure there is no time wasted when waiting.
 * Can be configured to randomly sleep for random duration (customizable) as well as add additional delay (customizable) between clicks to behave more human-like.

# Usage

1. Create tob.lt account. **See bottom of this page in order to avoid your other accounts getting banned!**.

Perform these settings in order to ensure smooth experience:
* `Ijungti/Isjungti leidima siulyti mainus` - disabled
* `Pranesimas apie barono atvykima` - disabled
* `Pranesimas apie feja` - disabled
* `Grafiniu ikonu ir paveiksleliu rodymas` - disabled
* `Greitasis meniu` - leave only `Antiflood laikrodukas` and hide the rest.
* `Zaidimo fonas` - disabled

2. Create Telegram bot: https://core.telegram.org/bots

Set below commands for your bot (using BotFather):
```
start - Resume bot
stop - Pause bot
```

**Tip from my experience**: Change Telegram bot's notification sound. In my case it was possible in (official/standard) Telegram app --> settings. Otherwise you will miss it at some point. :)

2. [Install Golang](https://golang.org/doc/install).

Most of the "trusted" distros ships a very old Golang version in their official repositories, which might not work at all. Make sure to remove any existing Golang installations and install the latest version using [official upstream guide](https://golang.org/doc/install) for your operating system.

3. Install dependencies

```
# Runtime
tesseract
tesseract Lithuanian pack

# Compiling
tesseract development package
gcc
g++
```

Examples:
```bash
# Ubuntu/Debian
apt install tesseract-ocr tesseract-ocr-lit libtesseract-dev gcc g++

# Fedora/RHEL
dnf install tesseract tesseract-langpack-lit tesseract-devel gcc g++
```

4. Build binary
```
# Build binary
go build -ldflags="-s -w" -o tobot ./cmd/tobot/main.go

# Verify binary
./tobot -help
```

5. Create configuration file. Simply copy `config.example.yml` to a new file `config.yml` and edit accordingly.

Telegram bot will notify you about important events (e.g. bot started or you got banned).

Telegram bot will also send you all received new PMs from the players. Reply to the player by simply **replying** to the same Telegram **bot's message** (not just sending message to bot). Note that tob.lt bot **WILL STOP indefinitely** until you reply to the player. If you don't want to reply to the player, then reply to Telegram bot's message with text `/ignore`. Also note that Telegram bot will not send any message to the player that starts with `/`, so it's OK to make a TYPO (mistake) that starts with `/`).

6. Create new directory, similar to existing one `activities/*` (use those dirs as an example). Each file represents different activity, format must be `*.yml` and such files will be executed in alphabetical filename order (hence that's the meaning of `10_` in filenames). Once all activities are finished, bot will start from the top again. :)

Full list of modules: https://github.com/erkexzcx/tobot/tree/master/module

Non-module specific fields:
```
_module - (required) name of the module
_count - (optional) how many times perform the module action. 
```

All other fields are listed in README.md file within each module's directory.

**Note**: Feel free to use `activities/*` as they are premade templates. E.g. `activities/day1` works just fine on fresh account (do not forget to look at `activities/day1/reikalavimai.txt`).

7. Run program
```
./tobot -help

./tobot
./tobot -config /path/to/config
```

# Notes/Tips regarding tob.lt and this bot
  - Moderators constantly check via PM if you are human, e.g. "Tikrinu 8512v. atrasyk: Kelmas" or "Tikrinu 9999v kiek bus 5+2?" and something like this. Failure to reply within minutes will result in ban (and likelly account deletion). There is no reliable method to automate replies, even tho I had some success.. :)
  - Moderators can see your PM, so moving/trading between multiple accounts is not a solution. Not sure about "Siukslynes" - throwing away and picking using other account.
  - Moderators would ban you if you level up only one level at a time. Without a warning of course...
  - Moderators would ban you if you level up (all levels) for prolonged period of time. I've got ban & account deleted after non stop clicking for ~28h and daily clicks record was about 14k. xD And no, replying to each message of moderator does not guarantee that you won't be banned. :D
  - All accounts are storing your IPs, so getting your single account removed might remove your other accounts as well.
  - One moderator approached me with statement that he/she monitored my click intervals and they were identical. `tobot` does support randomized waiting intervals between clicks. I am not aware if moderators can see if you go offline between actioning (`become_offline` option), but it makes sense to use this config.

# Tips on running 24/7 (sort of)

Start by using below configuration fragment which works great:
```yaml
settings:
  # Do not stay online 24/7
  become_offline:
    enabled: false # disabled
    every: 1h,2h
    for: 30m,60m

  # Add additional random delay between clicks
  randomize_wait:
    enabled: true
    wait_val: 0ms,4000ms
```

Then setup SystemD service `/etc/systemd/system/tobot.service` as per example below:
```
[Unit]
Description=tobot service
After=network-online.target

[Service]
User=erikas
Group=erikas
WorkingDirectory=/home/erikas/tobot
ExecStart=/home/erikas/tobot/tobot -config config.yml
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
```

Then enable/start it:
```
systemctl daemon-reload
systemctl enable --now tobot.service
```

Lastly, setup cronjob to start & stop this bot at 9AM and stop at 9PM:
```
# tob.lt bot
0 7 * * * systemctl start tobot.service
0 19 * * * systemctl stop tobot.service
```

By using above setup, bot does around ~5200 (+-300) clicks per day, which is around 5-8 place of the top clickers per day. By not being #1, moderators do not spam you whether you are human, so this is fairly safe amount of clicks per day. :)
