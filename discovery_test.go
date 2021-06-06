package botnet_test

import (
	"reflect"
	"testing"

	"github.com/ghostec/botnet"
)

func TestDiscovery(t *testing.T) {
	dsc := botnet.NewDiscovery()

	defer dsc.Stop()
	go dsc.Start("localhost", 8333)

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

func TestDiscoveryWhenBotIsOffline(t *testing.T) {
	dsc := botnet.NewDiscovery()

	defer dsc.Stop()
	go dsc.Start("localhost", 8333)

	dsc.Wait()

	if !reflect.DeepEqual([]string{"botnet"}, dsc.Bots()) {
		t.Fatal("discovery.Bots() isn't []string('botnet')")
	}

	bot := botnet.NewBot("ghostec")

	if err := bot.Connect("localhost", 8333); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]string{"botnet", "ghostec"}, dsc.Bots()) {
		t.Fatal("discovery.Bots() isn't []string('botnet')")
	}

	if err := bot.Stop(); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]string{"botnet"}, dsc.Bots()) {
		t.Fatal("discovery.Bots() isn't []string('botnet')")
	}
}

type sstring string

func (str *sstring) Unmarshal(b []byte) error {
	*str = sstring(b)
	return nil
}
