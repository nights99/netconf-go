// +build wasm

// GOOS=js GOARCH=wasm go build -o main.wasm
// ~/go/bin/goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'
// cp main.wasm xxx

package main

import (
	"fmt"
	"syscall/js"
)

func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		modNames := getSchemaList(globalSession)
		new := make([]interface{}, len(modNames))
		for i, v := range modNames {
			new[i] = v
		}
		return new
	})
	return jsonFunc
}

func main() {
	// Connect("localhost", 12345)
	globalSession, _ = DialWebSocket()
	// modNames := GetModNames()
	// fmt.Printf("Mod names: %v\n", modNames)
	js.Global().Set("formatJSON", jsonWrapper())
	// js.Global().Set("getEntries", getEntries())
	<-make(chan bool)

}
