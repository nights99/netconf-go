package main

import (
	"context"
	"fmt"
	"io"
	"log"

	// "syscall/js"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"nhooyr.io/websocket"
)

var cancel context.CancelFunc
var ctx context.Context

// TransportWebSocket x
type TransportWebSocket struct {
	netconf.TransportBasicIO
	wsConn *websocket.Conn
}

// Dial x
func (t *TransportWebSocket) Dial() error {
	t.wsConn = Connect("foo", 1234)

	t.ReadWriteCloser = t
	return nil
}

func (t *TransportWebSocket) Read(p []byte) (int, error) {
	println("Read called")
	return 0, io.EOF
}

func (t *TransportWebSocket) Write(p []byte) (int, error) {
	println("Write called")
	return 1, nil
}

func (t *TransportWebSocket) Close() error {
	println("Close called")
	return nil
}

// Connect ...
func Connect(address string, port int) *websocket.Conn {
	var err error
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute)

	ws, _, err := websocket.Dial(ctx, "ws://localhost:12345", nil)
	if err != nil {
		log.Fatal("Failed to connect: ", err)
	}

	fmt.Printf("Connected to %s:%d\n", address, port)
	ws.SetReadLimit(32 * 1024 * 4)
	return ws
}

// DialWebSocket x
func DialWebSocket() (*netconf.Session, error) {
	var t TransportWebSocket
	err := t.Dial()
	if err != nil {
		// t.Close()
		return nil, err
	}
	return netconf.NewSession(&t), nil
}
