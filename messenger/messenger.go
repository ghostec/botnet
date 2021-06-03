package messenger

type Messenger interface {
	Start() error
	Stop() error
	Send([]byte) error
	Recv() ([]byte, error)
	Ask([]byte) ([]byte, error)
	OnMessage(Handler)
	OnNewClient(Messenger) error
}
