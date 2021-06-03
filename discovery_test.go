package botnet_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ghostec/botnet"
)

func TestDiscovery(t *testing.T) {
	dsc := botnet.NewDiscovery()

	go func() {
		if err := dsc.Start("localhost", 8333); err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	bot := botnet.NewBot("ghostec")

	if err := bot.Connect("localhost", 8333); err != nil {
		t.Fatal(err)
	}

	b, err := bot.Ask("botnet", "echo", []byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b, []byte("hello world")) {
		t.Fatal("expected 'hello world', got " + string(b))
	}
}
