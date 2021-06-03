package botnet_test

import (
	"testing"

	"github.com/ghostec/botnet"
)

func TestDiscovery(t *testing.T) {
	dsc := botnet.NewDiscovery()

	go func() {
		if err := dsc.Start("localhost", 8333); err != nil {
			t.Error(err)
		}
	}()

	dsc.Wait()

	bot := botnet.NewBot("ghostec")

	if err := bot.Connect("localhost", 8333); err != nil {
		t.Fatal(err)
	}

	var str sstring
	if err := bot.Ask("botnet", "echo", []byte("hello world")).To(&str); err != nil {
		t.Fatal(err)
	}

	if str != "hello world" {
		t.Fatal("expected 'hello world', got " + str)
	}
}

type sstring string

func (str *sstring) Unmarshal(b []byte) error {
	*str = sstring(b)
	return nil
}
