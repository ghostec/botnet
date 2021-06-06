package botnet_test

import (
	"testing"

	"github.com/ghostec/botnet"
)

func BenchmarkBotAsk(b *testing.B) {
	dsc := botnet.NewDiscovery()

	defer dsc.Stop()

	go dsc.Start("localhost", 8333)

	dsc.Wait()

	bot := botnet.NewBot("ghostec")

	if err := bot.Connect("localhost", 8333); err != nil {
		b.Fatal(err)
	}

	var str sstring

	for n := 0; n < b.N; n++ {
		if err := bot.Ask("botnet", "echo", []byte("a")).To(&str); err != nil {
			b.Fatal(err)
		}
	}
}
