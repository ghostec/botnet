package messenger

type Handler func([]byte) ([]byte, error)
