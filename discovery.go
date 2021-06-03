package botnet

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Discovery struct {
	nextInvokeID uint64
	didListen    chan bool

	mbots sync.Mutex
	bots  map[string][]*websocket.Conn

	manswers sync.Mutex
	answers  map[uint64]chan []byte
}

func NewDiscovery() *Discovery {
	return &Discovery{
		didListen: make(chan bool, 1),
		bots:      map[string][]*websocket.Conn{},
		answers:   map[uint64]chan []byte{},
	}
}

func (dsc *Discovery) Start(host string, port int) error {
	errch := make(chan error)

	go func() {
		if err := dsc.startServer(host, port); err != nil {
			errch <- err
		}
	}()

	select {
	case <-dsc.didListen:
	case err := <-errch:
		return err
	}

	go func() {
		if err := dsc.startBot(host, port); err != nil {
			errch <- err
		}
	}()

	return <-errch
}

func (dsc *Discovery) startServer(host string, port int) error {
	mux := http.NewServeMux()

	var upgrader = websocket.Upgrader{} // use default options

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
		}

		defer c.Close()

		_, b, err := c.ReadMessage()
		if err != nil {
			log.Println("can't read first message")
			return
		}

		dsc.mbots.Lock()
		dsc.bots[string(b)] = append(dsc.bots[string(b)], c)
		dsc.mbots.Unlock()

		for {
			_, b, err := c.ReadMessage()
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
				if err := dsc.invokeanswer(c, b); err != nil {
					log.Println("invokeanswer:", err)
				}
			case byte('2'):
				if err := dsc.ask(c, b); err != nil {
					log.Println("ask:", err)
				}
			default:
				log.Println("unknown message type")
			}
		}
		return
	})

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	dsc.didListen <- true

	return http.Serve(l, mux)
}

func (dsc *Discovery) invokeanswer(conn *websocket.Conn, iab []byte) error {
	var ia invokeanswer
	if err := ia.Unmarshal(iab); err != nil {
		return nil
	}

	dsc.manswers.Lock()
	ch, ok := dsc.answers[ia.InvokeID]
	if !ok {
		return fmt.Errorf("invalid invoke id %d", ia.InvokeID)
	}
	dsc.manswers.Unlock()

	ch <- ia.Content

	return nil
}

func (dsc *Discovery) ask(conn *websocket.Conn, askb []byte) error {
	var ask ask
	if err := ask.Unmarshal(askb); err != nil {
		return nil
	}

	bots, ok := dsc.bots[ask.BotName]
	if !ok {
		return nil
	}

	if len(bots) == 0 {
		return nil
	}

	dsc.manswers.Lock()
	invokeID := dsc.nextInvokeID
	dsc.nextInvokeID += 1
	ch := make(chan []byte, 1)
	dsc.answers[invokeID] = ch
	dsc.manswers.Unlock()

	defer func() {
		dsc.manswers.Lock()
		delete(dsc.answers, invokeID)
		dsc.manswers.Unlock()
	}()

	b, err := invoke{ID: invokeID, Action: ask.Action, Content: ask.Content}.Marshal()
	if err != nil {
		return err
	}

	if err := bots[0].WriteMessage(websocket.BinaryMessage, b); err != nil {
		return err
	}

	b = <-ch

	b, err = answer{AskID: ask.ID, Content: b}.Marshal()
	if err != nil {
		return nil
	}

	return conn.WriteMessage(websocket.BinaryMessage, b)
}

func (dsc *Discovery) startBot(host string, port int) error {
	bot := NewBot("botnet")

	if err := bot.Connect("localhost", 8333); err != nil {
		return err
	}

	bot.Handle("echo", func(content []byte) ([]byte, error) {
		return content, nil
	})

	return nil
}
