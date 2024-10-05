//go:build wasm

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"

	// "syscall/js"
	"time"

	// netconf "github.com/nemith/go-netconf/v2"
	// "github.com/nemith/go-netconf/v2/transport"
	netconf "github.com/nemith/netconf"
	"github.com/nemith/netconf/transport"
	"nhooyr.io/websocket"
)

//lint:ignore U1000 x
var cancel context.CancelFunc
var ctx context.Context

// TransportWebSocket x
type TransportWebSocket struct {
	wsConn  *websocket.Conn
	lastMsg []byte
	offset  int
	*transport.Framer
}

// Dial x
func (t *TransportWebSocket) Dial(address string, port int) error {
	t.wsConn, _ = Connect(address, port)
	t.Framer = transport.NewFramer(t, t)

	return nil
}

func (t *TransportWebSocket) Read(p []byte) (int, error) {
	// println("Read called")
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if t.lastMsg == nil {
		t.offset = 0
		// var mt websocket.MessageType
		var err error
		_, t.lastMsg, err = t.wsConn.Read(ctx)
		if err != nil {
			log.Printf("Ws read err: %v\n", err)
			return 0, err
		}
		// log.Printf("Ws read: %v\n", string(t.lastMsg))
	}

	var bytesCopied int = 0
	if len(t.lastMsg)-t.offset > 0 {
		bytesCopied = copy(p, t.lastMsg[t.offset:])
	}

	// log.Printf("Wsread: %d %d %d\n", len(t.lastMsg), t.offset, len(p))
	if len(t.lastMsg)-t.offset > len(p) {
		t.offset += bytesCopied
		// log.Printf("Ws read: return %d\n", bytesCopied)
		return bytesCopied, nil
	} else if len(t.lastMsg)-t.offset > 0 {
		t.offset += bytesCopied
		// log.Printf("Ws read: return !EOF %d %d %v\n", bytesCopied, len(t.lastMsg)-t.offset, t.lastMsg[t.offset-bytesCopied:])
		// log.Printf("Ws read: return !EOF %d %d\n", bytesCopied, len(t.lastMsg)-t.offset)
		t.lastMsg = nil
		t.offset = 0
		return bytesCopied, nil
	}
	t.lastMsg = nil
	t.offset = 0
	// log.Printf("Ws read: return EOF\n")

	return 0, io.EOF
}

func (t *TransportWebSocket) Write(p []byte) (int, error) {
	// log.Printf("Write called %d bytes %v\n", len(p), string(p))

	ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := t.wsConn.Write(ctx, websocket.MessageBinary, p)
	if err != nil {
		log.Printf("Ws write err: %v\n", err)
		return 0, err
	}

	return len(p), nil
	// return 1, nil
}

// Close x
func (t *TransportWebSocket) Close() error {
	println("Close called")
	return nil
}

// Connect ...
func Connect(address string, port int) (*websocket.Conn, error) {
	var err error
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ws, _, err := websocket.Dial(ctx, "ws://"+address+":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Panic("Failed to connect: ", err)
		return nil, err
	}

	fmt.Printf("Connected to %s:%d\n", address, port)
	ws.SetReadLimit(1024 * 1024)
	return ws, nil
}

// DialWebSocket x
func DialWebSocket(address string, port int) (*netconf.Session, error) {
	var t TransportWebSocket
	err := t.Dial(address, port)
	if err != nil {
		// t.Close()
		return nil, err
	}
	session, err := netconf.Open(&t)
	if err != nil {
		panic(err)
	}
	return session, nil
}
