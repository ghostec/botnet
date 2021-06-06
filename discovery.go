package botnet

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type Discovery struct {
	nextInvokeID uint64
	didListen    chan bool
	ready        chan bool

	mbots sync.Mutex
	bots  map[string][]*websocket.Conn

	manswers sync.Mutex
	answers  map[uint64]chan []byte
}

func NewDiscovery() *Discovery {
	return &Discovery{
		ready:     make(chan bool, 1),
		didListen: make(chan bool, 1),
		bots:      map[string][]*websocket.Conn{},
		answers:   map[uint64]chan []byte{},
	}
}

func (dsc *Discovery) Wait() {
	<-dsc.ready
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

	time.Sleep(1 * time.Second)

	if err := dsc.startBot(host, port); err != nil {
		return err
	}

	close(dsc.ready)

	return <-errch
}

func (dsc *Discovery) startServer(host string, port int) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Println("upgrade:", err)
		}

		c.SetReadLimit(10485760)

		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		_, b, err := c.Read(context.Background())
		if err != nil {
			log.Println("can't read first message")
			return
		}

		dsc.mbots.Lock()
		dsc.bots[string(b)] = append(dsc.bots[string(b)], c)
		dsc.mbots.Unlock()

		println(string(b))

		for {
			_, b, err := c.Read(context.Background())
			if err != nil {
				log.Println("read:", err)
				break
			}

			if len(b) == 0 {
				log.Println("empty message")
				break
			}

			println(string(b))

			go func() {
				// 0 = unknown, 1 = invoke, 2 = answer
				switch b[0] {
				case byte('1'):
					println("invokeanswer")
					if err := dsc.invokeanswer(c, b); err != nil {
						log.Println("invokeanswer:", err)
					}
				case byte('2'):
					println("ask", string(b))
					if err := dsc.ask(c, b); err != nil {
						log.Println("ask:", err)
					}
				default:
					log.Println("unknown message type")
				}
			}()
		}
		return
	})

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	close(dsc.didListen)

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

	println("ask invoke")

	if err := bots[int(invokeID)%len(bots)].Write(context.Background(), websocket.MessageBinary, b); err != nil {
		return err
	}

	b = <-ch

	b, err = answer{AskID: ask.ID, Content: b}.Marshal()
	if err != nil {
		return nil
	}

	return conn.Write(context.Background(), websocket.MessageBinary, b)
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
