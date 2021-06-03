package botnet_test

import (
	"testing"

	"github.com/ghostec/botnet"
	"github.com/ghostec/botnet/messenger/websocket"
)

func TestBot(t *testing.T) {
	m0 := websocket.New(websocket.ServerOption("localhost", 8333))
	b0 := botnet.NewBot("botnet", m0)

	b0.Handle("echo", func(str string) string {
		return str
	})

	b0.Start()

	m1 := websocket.New(websocket.DialOption("localhost", 8333))
	b1 := botnet.NewBot("ghostec", m1)
	b1.Start()

	var str string
	if err := b1.Ask("botnet", "echo", bytes([]byte("hello world"))).To(&str); err != nil {
		t.Fatal(err)
	}

	if str != "hello world" {
		t.Fatal("unexpected answer")
	}
}

type bytes []byte

func (b bytes) Marshal() ([]byte, error) {
	return []byte(b), nil
}
