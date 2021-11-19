package player

import tb "gopkg.in/tucnak/telebot.v2"

func (p *Player) NotifyTelegram(msg string) {
	p.telegramBot.Send(p.telegramChat, msg, &tb.SendOptions{})
}

func (p *Player) NotifyTelegramSilent(msg string) {
	p.telegramBot.Send(p.telegramChat, msg, &tb.SendOptions{DisableNotification: true})
}
