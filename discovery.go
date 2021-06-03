package botnet

import "github.com/ghostec/botnet/messenger"

type Discovery struct {
	bot *Bot
}

func NewDiscovery(bot *Bot) *Discovery {
	bot.Messenger().OnNewClient(func(clientMessenger messenger.Messenger) error {
		clientBot := NewBot(clientMessenger)
		clientBot.Ask()
	})

	return &Discovery{bot: bot}
}

func (dsc *Discovery) Start() error {
	return dsc.bot.start()
}

func (dsc *Discovery) Stop() error {
	return dsc.bot.Stop()
}

func (dsc *Discovery) onMessage() ([]byte, error) {
	return nil, nil
}
