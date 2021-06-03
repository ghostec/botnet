package websocket_test

import (
	"bytes"
	"testing"

	"github.com/ghostec/botnet/messenger/websocket"
)

func TestWebSocketSendRecv(t *testing.T) {
	m0 := websocket.New(websocket.ServerOption("localhost", 8333))

	m0.OnMessage(func(b []byte) ([]byte, error) {
		return b, nil
	})

	m0.Start()
	defer m0.Stop()

	m1 := websocket.New(websocket.DialOption("localhost", 8333))
	m1.Start()

	if err := m1.Send([]byte("hello world")); err != nil {
		t.Fatal(err)
	}

	b, err := m0.Recv()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b, []byte("hello world")) {
		t.Fatal("unexpected answer")
	}
}

func TestWebSocketAsk(t *testing.T) {
	m0 := websocket.New(websocket.ServerOption("localhost", 8333))

	m0.OnMessage(func(b []byte) ([]byte, error) {
		return b, nil
	})

	m0.Start()
	defer m0.Stop()

	m1 := websocket.New(websocket.DialOption("localhost", 8333))
	m1.Start()

	b, err := m1.Ask([]byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b, []byte("hello world")) {
		t.Fatal("unexpected answer")
	}
}
