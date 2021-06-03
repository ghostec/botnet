package botnet

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/cenkalti/backoff"
	"github.com/gorilla/websocket"
)

type Bot struct {
	name      string
	conn      *websocket.Conn
	nextAskID uint64

	mhandlers sync.Mutex
	handlers  map[string]func([]byte) ([]byte, error)

	manswers sync.Mutex
	answers  map[uint64]chan []byte
}

func NewBot(name string) *Bot {
	return &Bot{
		name:     name,
		answers:  map[uint64]chan []byte{},
		handlers: map[string]func([]byte) ([]byte, error){},
	}
}

func (bot *Bot) Connect(host string, port int) error {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", host, port), Path: "/"}
	log.Printf("connecting to %s", u.String())

	if err := backoff.Retry(func() error {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			return err
		}

		bot.conn = c

		return nil
	}, backoff.NewExponentialBackOff()); err != nil {
		return err
	}

	if err := bot.conn.WriteMessage(websocket.BinaryMessage, []byte(bot.name)); err != nil {
		return err
	}

	go func() {
		for {
			_, b, err := bot.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}

			if len(b) == 0 {
				log.Println("empty message")
				break
			}

			// 0 = unknown, 1 = invoke, 2 = answer
			switch b[0] {
			case byte('1'):
				if err := bot.invoke(b); err != nil {
					log.Println("invoke:", err)
				}
			case byte('2'):
				if err := bot.answer(b); err != nil {
					log.Println("answer:", err)
				}
			default:
				log.Println("unknown message type")
			}
		}
	}()

	return nil
}

func (bot *Bot) answer(content []byte) error {
	var answer answer
	if err := answer.Unmarshal(content); err != nil {
		return err
	}

	bot.manswers.Lock()
	bot.answers[answer.AskID] <- answer.Content
	bot.manswers.Unlock()

	return nil
}

func (bot *Bot) invoke(content []byte) error {
	var invoke invoke

	if err := invoke.Unmarshal(content); err != nil {
		return err
	}

	bot.mhandlers.Lock()
	handler, ok := bot.handlers[invoke.Action]
	bot.mhandlers.Unlock()

	if !ok {
		return fmt.Errorf("no handler for %s", invoke.Action)
	}

	b, err := handler(invoke.Content)
	if err != nil {
		return err
	}

	b, err = invokeanswer{InvokeID: invoke.ID, Content: b}.Marshal()
	if err != nil {
		return err
	}

	return bot.conn.WriteMessage(websocket.BinaryMessage, b)
}

func (bot *Bot) Ask(botname, action string, content []byte) ([]byte, error) {
	bot.manswers.Lock()
	askID := bot.nextAskID
	bot.nextAskID += 1
	ch := make(chan []byte, 1)
	bot.answers[askID] = ch
	bot.manswers.Unlock()

	defer func() {
		bot.manswers.Lock()
		delete(bot.answers, askID)
		bot.manswers.Unlock()
	}()

	b, err := ask{ID: askID, BotName: botname, Action: action, Content: content}.Marshal()
	if err != nil {
		return nil, err
	}

	if err := bot.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return nil, err
	}

	return <-ch, nil
}

func (bot *Bot) Handle(action string, handler func([]byte) ([]byte, error)) {
	bot.handlers[action] = handler
}
