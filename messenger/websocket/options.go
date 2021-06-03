package websocket

type Option interface {
	apply(*WebSocket)
}

func ServerOption(host string, port int) serverOption {
	return serverOption{host: host, port: port}
}

type serverOption struct {
	host string
	port int
}

func (o serverOption) apply(ws *WebSocket) {
	ws.host = &o.host
	ws.port = &o.port
}

func DialOption(host string, port int) dialOption {
	return dialOption{host: host, port: port}
}

type dialOption struct {
	host string
	port int
}

func (o dialOption) apply(ws *WebSocket) {
	ws.dialHost = &o.host
	ws.dialPort = &o.port
}
