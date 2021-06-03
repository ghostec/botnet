package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/mathetake/gasm/wasi"
	"github.com/mathetake/gasm/wasm"
)

func main() {
	wasmBytes, _ := ioutil.ReadFile("../bot/lib.wasm")
	mod, err := wasm.DecodeModule(bytes.NewBuffer(wasmBytes))
	if err != nil {
		panic(err)
	}

	vm, err := wasm.NewVM(mod, wasi.New().Modules())
	if err != nil {
		panic(err)
	}

	ret, _, _ := vm.ExecExportedFunction("main")

	fmt.Println(ret)
}
