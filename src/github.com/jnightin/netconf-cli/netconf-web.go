// +build wasm

// GOOS=js GOARCH=wasm go build -o main.wasm
// ~/go/bin/goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'
// cp main.wasm xxx

package main

import (
	"fmt"
	"syscall/js"
)

var modNames []string

func getEntries() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// if len(args) != 1 {
		// 	fmt.Printf("Invalid no of arguments passed: %v\n", args)
		// 	return "Invalid no of arguments passed"
		// }
		fmt.Printf("input %v\n", args)
		slice := make([]string, len(args))
		for i, v := range args {
			slice[i] = v.String()
		}

		// GetEntries(slice)
		return nil
	})
	return jsonFunc
}

func doGetSchemas() {
	modNames = getSchemaList(globalSession)
	js.Global().Call("foo", modNames)
}

func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		// @@@ Can't block so may have to use 'go' for goroutine
		go doGetSchemas()
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
