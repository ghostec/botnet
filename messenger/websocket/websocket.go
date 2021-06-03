package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/cenkalti/backoff"
	"github.com/ghostec/botnet/messenger"
	"github.com/gorilla/websocket"
)

type WebSocket struct {
	host, dialHost     *string
	port, dialPort     *int
	didListen, didDial chan bool
	conn               *websocket.Conn
	recv               chan []byte
	nextAskID          uint64
	listener           net.Listener

	manswers sync.Mutex
	answers  map[uint64]chan []byte

	handler     func([]byte) ([]byte, error)
	onNewClient func(messenger.Messenger) error
}

func New(opts ...Option) *WebSocket {
	ws := &WebSocket{
		recv:      make(chan []byte),
		answers:   map[uint64]chan []byte{},
		didListen: make(chan bool, 1),
		didDial:   make(chan bool, 1),
	}

	for _, option := range opts {
		option.apply(ws)
	}

	return ws
}

func (ws *WebSocket) OnMessage(handler messenger.Handler) {
	ws.handler = handler
}

func (ws *WebSocket) OnNewClient(messenger messenger.Messenger) {
	ws.handler = handler
}

// TODO: make sync
func (ws *WebSocket) Start() error {
	if ws.host != nil && ws.port != nil {
		if ws.handler == nil {
			return errors.New("no handler registered")
		}

		go func() {
			if err := ws.startServer(); err != nil {
				log.Println("startServer:", err)
			}
		}()

		<-ws.didListen
	}

	if ws.dialHost != nil && ws.dialPort != nil {
		go func() {
			if err := ws.startClient(); err != nil {
				log.Println("startClient:", err)
			}
		}()

		<-ws.didDial
	}

	return nil
}

func (ws *WebSocket) startServer() error {
	mux := http.NewServeMux()

	var upgrader = websocket.Upgrader{} // use default options

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
		}

		defer c.Close()

		for {
			_, b, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}

			var ask ask
			var ans []byte

			if err := ask.Unmarshal(b); err == nil {
				bans, err := ws.handler(ask.Content)
				if err != nil {
					log.Println("handler:", err)
					break
				}

				ans, err = answer{AskID: ask.ID, Content: bans}.Marshal()
				if err != nil {
					log.Println("marshal:", err)
					break
				}
			} else {
				ans, err = ws.handler(ask.Content)
				if err != nil {
					log.Println("handler:", err)
					break
				}
			}

			err = c.WriteMessage(websocket.BinaryMessage, ans)
			if err != nil {
				log.Println("write:", err)
				break
			}

			ws.recv <- b
		}
		return
	})

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *ws.host, *ws.port))
	if err != nil {
		return err
	}

	ws.listener = l

	ws.didListen <- true

	return http.Serve(l, mux)
}

func (ws *WebSocket) Stop() error {
	return ws.listener.Close()
}

func (ws *WebSocket) startClient() error {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", *ws.dialHost, *ws.dialPort), Path: "/"}
	log.Printf("connecting to %s", u.String())

	if err := backoff.Retry(func() error {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			return err
		}

		ws.conn = c

		return nil
	}, backoff.NewExponentialBackOff()); err != nil {
		return err
	}

	ws.didDial <- true

	go func() {
		for {
			_, b, err := ws.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var ans answer
			if err := ans.Unmarshal(b); err == nil {
				ws.manswers.Lock()
				ch := ws.answers[ans.AskID]
				ws.manswers.Unlock()

				ch <- ans.Content
			}

			ws.recv <- b
		}
	}()

	return nil
}

func (ws *WebSocket) Send(b []byte) error {
	return ws.conn.WriteMessage(websocket.BinaryMessage, b)
}

func (ws *WebSocket) Recv() ([]byte, error) {
	b := <-ws.recv
	return b, nil
}

func (ws *WebSocket) Ask(b []byte) ([]byte, error) {
	askID := atomic.AddUint64(&ws.nextAskID, 1)

	ws.manswers.Lock()
	ch := make(chan []byte, 1)
	ws.answers[askID] = ch
	ws.manswers.Unlock()

	defer func() {
		ws.manswers.Lock()
		delete(ws.answers, askID)
		ws.manswers.Unlock()
	}()

	bask, err := ask{ID: askID, Content: b}.Marshal()
	if err != nil {
		return nil, err
	}

	if err := ws.Send(bask); err != nil {
		return nil, err
	}

	return <-ch, nil
}

type ask struct {
	ID      uint64 `json:"i"`
	Content []byte `json:"c"`
}

func (ask ask) Marshal() ([]byte, error) {
	return json.Marshal(ask)
}

func (ask *ask) Unmarshal(b []byte) error {
	return json.Unmarshal(b, ask)
}

type answer struct {
	AskID   uint64 `json:"a"`
	Content []byte `json:"c"`
}

func (answer answer) Marshal() ([]byte, error) {
	return json.Marshal(answer)
}

func (answer *answer) Unmarshal(b []byte) error {
	return json.Unmarshal(b, answer)
}
