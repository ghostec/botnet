package botnet

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ghostec/botnet/messenger"
)

type Bot struct {
	name      string
	messenger messenger.Messenger

	mhandlers sync.Mutex
	handlers  map[string]func(string) string
}

func NewBot(name string, messenger messenger.Messenger) *Bot {
	bot := &Bot{
		name:      name,
		messenger: messenger,
		handlers:  map[string]func(string) string{},
	}

	bot.Handle("botnet/discovery/info", func(_ string) string {
		return name
	})

	return bot
}

func (bot *Bot) Messenger() messenger.Messenger {
	return bot.Messenger
}

func (bot *Bot) Start() error {
	bot.messenger.OnMessage(func(b []byte) ([]byte, error) {
		var ask ask
		if err := ask.Unmarshal(b); err != nil {
			return nil, err
		}

		bot.mhandlers.Lock()
		handler, ok := bot.handlers[ask.Action]
		if !ok {
			bot.mhandlers.Unlock()
			return nil, fmt.Errorf("no handler: %s", ask.Action)
		}
		bot.mhandlers.Unlock()

		str := handler(string(ask.Content))
		return []byte(str), nil
	})

	return bot.messenger.Start()
}

func (bot *Bot) Handle(action string, handler func(str string) string) {
	bot.mhandlers.Lock()
	defer bot.mhandlers.Unlock()

	bot.handlers[action] = handler
}

// Only works when talking to botnet/discovery
func (bot *Bot) Ask(botName, action string, content Marshaler) Answer {
	b, err := content.Marshal()
	if err != nil {
		return Answer{err: err}
	}

	ask := ask{Bot: botName, Action: action, Content: b}
	bask, err := ask.Marshal()
	if err != nil {
		return Answer{err: err}
	}

	ans, err := bot.messenger.Ask(bask)
	if err != nil {
		return Answer{err: err}
	}

	return Answer{content: ans}
}

func (bot *Bot) Call(action string, content Marshaler) Answer {
	b, err := content.Marshal()
	if err != nil {
		return Answer{err: err}
	}

	call := call{Action: action, Content: b}
	bcall, err := call.Marshal()
	if err != nil {
		return Answer{err: err}
	}

	ans, err := bot.messenger.Call(bask)
	if err != nil {
		return Answer{err: err}
	}

	return Answer{content: ans}
}

type ask struct {
	Bot     string `json:"b"`
	Action  string `json:"a"`
	Content []byte `json:"c"`
}

func (ask ask) Marshal() ([]byte, error) {
	return json.Marshal(ask)
}

func (ask *ask) Unmarshal(b []byte) error {
	return json.Unmarshal(b, ask)
}

type Answer struct {
	content []byte
	err     error
}

func (ans Answer) To(str *string) error {
	if ans.err != nil {
		return ans.err
	}

	*str = string(ans.content)
	return nil
}

type Marshaler interface {
	Marshal() ([]byte, error)
}
