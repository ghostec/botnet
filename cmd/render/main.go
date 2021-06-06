package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ghostec/botnet"
)

func main() {
	bot := botnet.NewBot("ghostec")

	if err := bot.Connect("localhost", 8333); err != nil {
		panic(err)
	}

	width := 1000
	ar := 16.0 / 9.0
	height := int(float64(width) / ar)

	var wg sync.WaitGroup

	parallelism := 6

	wg.Add(parallelism)

	for i := 0; i < parallelism; i++ {
		j := i

		go func() {
			defer wg.Done()

			var args struct {
				LineA int `json:"a"`
				LineB int `json:"b"`
			}

			args.LineA = j * height / parallelism
			args.LineB = (j + 1) * height / parallelism

			if j == parallelism-1 {
				args.LineB = height
			}

			argsb, err := json.Marshal(args)
			if err != nil {
				panic(err)
			}

			var b bbyte
			if err := bot.Ask("ray", "render", argsb).To(&b); err != nil {
				panic(err)
			}

			if err := os.WriteFile(fmt.Sprintf("blend_%d.png", j), b.bytes, 0755); err != nil {
				panic(err)
			}
		}()
	}

	wg.Wait()
}

type bbyte struct {
	bytes []byte
}

func (bb *bbyte) Unmarshal(b []byte) error {
	bb.bytes = b
	return nil
}
