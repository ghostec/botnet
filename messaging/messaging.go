package messaging

import "encoding/json"

type Ask struct {
	ID      uint64 `json:"i"`
	BotName string `json:"b"`
	Action  string `json:"a"`
	Content []byte `json:"c"`
}

func (ask Ask) Marshal() ([]byte, error) {
	b, err := json.Marshal(ask)
	if err != nil {
		return nil, err
	}

	return append([]byte{'2'}, b...), nil
}

func (ask *Ask) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], ask)
}

type Answer struct {
	AskID   uint64 `json:"a"`
	Content []byte `json:"c"`
}

func (answer Answer) Marshal() ([]byte, error) {
	b, err := json.Marshal(answer)
	if err != nil {
		return nil, err
	}

	return append([]byte{'2'}, b...), nil
}

func (answer *Answer) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], answer)
}

type Invoke struct {
	ID      uint64 `json:"i"`
	Action  string `json:"a"`
	Content []byte `json:"c"`
}

func (invoke Invoke) Marshal() ([]byte, error) {
	b, err := json.Marshal(invoke)
	if err != nil {
		return nil, err
	}

	return append([]byte{'1'}, b...), nil
}

func (invoke *Invoke) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], invoke)
}

type InvokeAnswer struct {
	InvokeID uint64 `json:"i"`
	Content  []byte `json:"b"`
}

func (ia InvokeAnswer) Marshal() ([]byte, error) {
	b, err := json.Marshal(ia)
	if err != nil {
		return nil, err
	}

	return append([]byte{'1'}, b...), nil
}

func (ia *InvokeAnswer) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], ia)
}
