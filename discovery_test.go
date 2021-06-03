package botnet_test

import (
	"testing"

	"github.com/ghostec/botnet/messenger/websocket"
)

func TestDiscovery(t *testing.T) {
	mdsc := websocket.New(websocket.ServerOption("localhost", 8333))
	bdsc := botnet.NewBot("discovery", mdsc)
	dsc := botnet.NewDiscovery(bdsc)
	dsc.Start()
	// defer dsc.Stop()

	// action: botnet/discovery/info
	// answers "name", for example, it's all we need to route requests for now

	m0 := websocket.New(websocket.DialOption("localhost", 8333))
	b0 := botnet.New("botnet", m0)

	b0.Handle("echo", func(str string) string {
		return str
	})

	b0.Start()
	// defer b0.Stop()

	m1 := websocket.New(websocket.DialOption("localhost", 8333))
	b1 := botnet.New("client", m1)
	b1.Start()
	// defer b1.Stop()

	b1.Ask("")
}
