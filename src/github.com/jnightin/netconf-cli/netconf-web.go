// +build wasm

// GOOS=js GOARCH=wasm go build -o main.wasm
// ~/go/bin/goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'
// cp main.wasm xxx

package main

import (
	"fmt"
	"log"
	"strings"
	"syscall/js"

	"github.com/openconfig/goyang/pkg/yang"
)

var modNames []string

func doGetEntries(slice []string) {
	entries := listYang(strings.Join(slice, " "))

	webEntries := make([]string, 0)
	webEntries = append(webEntries, "GetEntries")
	webEntries = append(webEntries, slice[1:]...)
	webEntries = append(webEntries, ":")

	// Now we need each entry at this directory level.
	for _, v := range entries {
		// fmt.Printf("listYang returned %v\n", v)
		x := strings.Split(v, " ")
		webEntries = append(webEntries, x[len(x)-1:]...)
	}

	js.Global().Call("foo", strings.Join(webEntries, " "))
}

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
		slice = append([]string{"get-oper"}, slice...)
		go doGetEntries(slice)
		return nil
	})
	return jsonFunc
}

func doGetSchemas() {
	modNames = getSchemaList(globalSession)
	log.Printf("Got schemas: %v\n", modNames[:3])
	// js.Global().Call("foo", modNames)
	js.Global().Call("foo", strings.Join(modNames, " "))
	// js.Global().Call("foo", "test string")
}

func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)

		go doGetSchemas()

		modNames := []string{"foo", "bar"}
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
	ms = yang.NewModules()
	globalSession, _ = DialWebSocket("jnightin-ads2.cisco.com", 12345)
	// modNames := GetModNames()
	// fmt.Printf("Mod names: %v\n", modNames)
	js.Global().Set("formatJSON", jsonWrapper())
	js.Global().Set("getEntries", getEntries())
	<-make(chan bool)

}
