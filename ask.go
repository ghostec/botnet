package botnet

import "encoding/json"

type ask struct {
	ID      uint64 `json:"i"`
	BotName string `json:"b"`
	Action  string `json:"a"`
	Content []byte `json:"c"`
}

func (ask ask) Marshal() ([]byte, error) {
	b, err := json.Marshal(ask)
	if err != nil {
		return nil, err
	}

	return append([]byte{'2'}, b...), nil
}

func (ask *ask) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], ask)
}

type answer struct {
	AskID   uint64 `json:"a"`
	Content []byte `json:"c"`
}

func (answer answer) Marshal() ([]byte, error) {
	b, err := json.Marshal(answer)
	if err != nil {
		return nil, err
	}

	return append([]byte{'2'}, b...), nil
}

func (answer *answer) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], answer)
}

type invoke struct {
	ID      uint64 `json:"i"`
	Action  string `json:"a"`
	Content []byte `json:"c"`
}

func (invoke invoke) Marshal() ([]byte, error) {
	b, err := json.Marshal(invoke)
	if err != nil {
		return nil, err
	}

	return append([]byte{'1'}, b...), nil
}

func (invoke *invoke) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], invoke)
}

type invokeanswer struct {
	InvokeID uint64 `json:"i"`
	Content  []byte `json:"b"`
}

func (invokeanswer invokeanswer) Marshal() ([]byte, error) {
	b, err := json.Marshal(invokeanswer)
	if err != nil {
		return nil, err
	}

	return append([]byte{'1'}, b...), nil
}

func (invokeanswer *invokeanswer) Unmarshal(b []byte) error {
	return json.Unmarshal(b[1:], invokeanswer)
}
