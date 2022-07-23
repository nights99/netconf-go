//go:build wasm
// +build wasm

// GOOS=js GOARCH=wasm go build -o main.wasm
// ~/go/bin/goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'
// cp main.wasm xxx

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"syscall/js"

	"github.com/openconfig/goyang/pkg/yang"
	log "github.com/sirupsen/logrus"
)

var modNames []string

var sessionLock sync.Mutex
var sessionCond *sync.Cond

func getEntry(this js.Value, args []js.Value) interface{} {
	log.Infoln("Go entry")
	fmt.Printf("getEntry input %v %v %v\n", this, args, len(args))
	yangClassName := args[0].Index(0).String()
	mod := ms.Modules[yangClassName]
	fmt.Printf("getEntryx %v %v\n", mod, ms)
	entry := yang.ToEntry(mod)

	for i := 1; i < args[0].Length(); i++ {
		v := args[0].Index(i)
		// log.Printf("Bar77: %v %v %v", v, i, entry)
		if v.String() == "" || entry == nil {
			break
		}
		entry = entry.Dir[v.String()]
	}

	foo, _ := json.Marshal(entry)

	bar := js.ValueOf(string(foo))

	return bar
}

func doGetEntries(slice []string) {
	if modNames == nil {
		modNames = getSchemaList(globalSession)
	}
	entries, _ := listYang(strings.Join(slice, " "))

	webEntries := make([]string, 0)
	webEntries = append(webEntries, "GetEntries")
	webEntries = append(webEntries, slice[1:]...)
	webEntries = append(webEntries, ":")

	// Now we need each entry at this directory level.
	for _, v := range entries {
		// fmt.Printf("listYang returned %v\n", v)
		x := strings.Split(v, " ")
		webEntries = append(webEntries, x...)
	}

	js.Global().Call("foo", strings.Join(webEntries, " "))
}

func getEntries(this js.Value, args []js.Value) interface{} {
	log.Infoln("Go entry")
	fmt.Printf("getEntries input %v\n", args)
	// sessionCond.L.Lock()
	// for globalSession == nil {
	// 	sessionCond.Wait()
	// }
	// sessionCond.L.Unlock()
	fmt.Printf("getEntries input2 %v %v %v %v\n", args, args[0].Type(), args[0].Length(), args[0].Index(0))
	slice := make([]string, args[0].Length())
	for i := 0; i < args[0].Length(); i++ {
		slice[i] = args[0].Index(i).String()
	}
	slice = append([]string{"get-oper"}, slice...)
	fmt.Printf("slice %v\n", slice)
	go doGetEntries(slice)
	return nil
}

func doGetSchemas(resolve *js.Value) {
	sessionCond.L.Lock()
	for globalSession == nil {
		sessionCond.Wait()
	}
	sessionCond.L.Unlock()
	log.Printf("Getting schemas\n")
	modNames = getSchemaList(globalSession)
	log.Printf("Got schemas: %v\n", modNames[:3])
	js.Global().Call("foo", strings.Join(modNames, " "))
	if resolve != nil {
		resolve.Invoke(strings.Join(modNames, " "))
	}
}

func GetModNames3(this js.Value, args []js.Value) interface{} {
	resolve := args[0]

	go doGetSchemas(&resolve)

	return nil
}

func jsonWrapper(this js.Value, args []js.Value) interface{} {
	log.Infoln("Go entry")
	if len(args) != 1 {
		return "Invalid no of arguments passed"
	}
	promise := js.Global().Get("Promise").New(js.FuncOf(GetModNames3))
	inputJSON := args[0].String()
	fmt.Printf("input-jsonWrapper %s\n", inputJSON)

	return promise
}

func sendNetconfRequest3(resolve *js.Value, req []string, reqType int) {
	netconfData, data := sendNetconfRequest(globalSession, strings.Join(req, " "), reqType)
	fmt.Printf("sendNetconfRequest3: %v, %v\n", netconfData, data)

	if resolve != nil {
		resolve.Invoke(data)
	}
}

func sendNetconfRequest1(this js.Value, args []js.Value) interface{} {
	log.Infoln("Go entry", args[1])
	slice := make([]string, args[0].Length())
	for i := 0; i < args[0].Length(); i++ {
		slice[i] = args[0].Index(i).String()
	}
	slice = append([]string{args[1].String()}, slice...)
	var reqType int
	switch args[1].String() {
	case "commit":
		reqType = commit
	default:
		reqType = getOper
	}

	promise := js.Global().Get("Promise").New(js.FuncOf(
		func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			go sendNetconfRequest3(&resolve, slice, reqType)
			return nil
		},
	))

	return promise
}

func connect(this js.Value, args []js.Value) interface{} {
	log.Infoln("Go entry")
	promise := js.Global().Get("Promise").New(js.FuncOf(
		func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			go func(resolve *js.Value) {
				var err error = nil
				globalSession, err = DialWebSocket("localhost", 12345)
				if err != nil {
					log.Panicf("%v\n", err)
				} else {
					fmt.Printf("Connected ok\n")
					sessionCond.Broadcast()
				}
				if resolve != nil {
					resolve.Invoke()
				}
			}(&resolve)
			return nil
		},
	))

	return promise

}

func main() {
	sessionCond = sync.NewCond(&sessionLock)
	js.Global().Set("formatJSON", js.FuncOf(jsonWrapper))
	js.Global().Set("getEntries", js.FuncOf(getEntries))
	js.Global().Set("getEntry", js.FuncOf(getEntry))
	js.Global().Set("sendNetconfRequest", js.FuncOf(sendNetconfRequest1))
	js.Global().Set("connect", js.FuncOf(connect))

	// log.SetLevel(log.InfoLevel)
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)

	// Connect("localhost", 12345)
	ms = yang.NewModules()
	// globalSession, _ = DialWebSocket("jnightin-ads2.cisco.com", 12345)

	// var err error = nil
	fmt.Printf("Before main connect\n")

	// globalSession, err = DialWebSocket("localhost", 12345)
	// if err != nil {
	// 	log.Panicf("%v\n", err)
	// } else {
	// 	fmt.Printf("Connected ok\n")
	// 	sessionCond.Broadcast()
	// }
	// modNames := GetModNames()
	// fmt.Printf("Mod names: %v\n", modNames)
	println("Before make")
	<-make(chan bool)
	println("After make")
}
