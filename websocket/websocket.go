package websocket

import (
	"context"
	"fmt"
	"log"
	"net/url"

	nws "nhooyr.io/websocket"
)

type Conn struct {
	conn *nws.Conn
}

func Dial(host string, port int) (*Conn, error) {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", host, port), Path: "/"}
	log.Printf("connecting to %s", u.String())

	var conn *nws.Conn

	conn, _, err := nws.Dial(context.Background(), u.String(), nil)
	if err != nil {
		return nil, err
	}

	conn.SetReadLimit(10485760)

	return &Conn{conn: conn}, nil
}

func (ws *Conn) ReadMessage() ([]byte, error) {
	_, b, err := ws.conn.Read(context.Background())
	return b, err
}

func (ws *Conn) WriteMessage(b []byte) error {
	return ws.conn.Write(context.Background(), nws.MessageBinary, b)
}
