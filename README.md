# Tobot

The [tob.lt](http://tob.lt/) bot, written in Go. 

Main features:
 * Automatically solves anti-bot captchas.
 * Automatically reply PM messages using OpenAI API.
 * Level up multiple skills at the same time using your very custom routines.
 * Supports entities like `Pragaro Vartai` and `Demonas`.
 * Multi-player support.
 * Automatically recover from network or `NUORODAS REIKIA SPAUSTI TIK VIENĄ KARTĄ!`-like errors.
 * Supports random sleep intervals and random additional wait between clicks (human-like clicking).
 * Maximum clicking performance - uses your provided RTT duration to compensate network latency between clicks.
 * Ability to specify separate root URL for each player, making possible to hide your IP using reverse-proxy.

Documentation: https://github.com/erkexzcx/tobot/wiki
