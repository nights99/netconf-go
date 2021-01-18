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

func getEntries(this js.Value, args []js.Value) interface{} {
	// if len(args) != 1 {
	// 	fmt.Printf("Invalid no of arguments passed: %v\n", args)
	// 	return "Invalid no of arguments passed"
	// }
	fmt.Printf("input %v\n", args)
	fmt.Printf("input %v %v %v %v\n", args, args[0].Type(), args[0].Length(), args[0].Index(0))
	slice := make([]string, args[0].Length())
	// for i, v := range args {
	// 	slice[i] = v.String()
	// }
	for i := 0; i < args[0].Length(); i++ {
		slice[i] = args[0].Index(i).String()
	}
	slice = append([]string{"get-oper"}, slice...)
	fmt.Printf("slice %v\n", slice)
	go doGetEntries(slice)
	return nil
}

func doGetSchemas(resolve *js.Value) {
	modNames = getSchemaList(globalSession)
	log.Printf("Got schemas: %v\n", modNames[:3])
	// js.Global().Call("foo", modNames)
	js.Global().Call("foo", strings.Join(modNames, " "))
	// js.Global().Call("foo", "test string")
	if resolve != nil {
		resolve.Invoke(strings.Join(modNames, " "))
	}
}

func GetModNames3(this js.Value, args []js.Value) interface{} {
	resolve := args[0]
	// modNames := []string{"shellutil", "ip-static"}
	// modNames := GetModNames2()

	go doGetSchemas(&resolve)

	// return modNames
	return nil
}

func jsonWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return "Invalid no of arguments passed"
	}
	promise := js.Global().Get("Promise").New(js.FuncOf(GetModNames3))
	inputJSON := args[0].String()
	fmt.Printf("input %s\n", inputJSON)

	// modNames := []string{"foo", "bar"}
	// new := make([]interface{}, len(modNames))
	// for i, v := range modNames {
	// 	new[i] = v
	// }
	// return new
	return promise
}

func main() {
	// Connect("localhost", 12345)
	ms = yang.NewModules()
	globalSession, _ = DialWebSocket("jnightin-ads2.cisco.com", 12345)
	// modNames := GetModNames()
	// fmt.Printf("Mod names: %v\n", modNames)
	js.Global().Set("formatJSON", js.FuncOf(jsonWrapper))
	js.Global().Set("getEntries", js.FuncOf(getEntries))
	<-make(chan bool)

}
