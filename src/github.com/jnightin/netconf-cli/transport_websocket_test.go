package main

import "testing"

func TestWS(t *testing.T) {
	s, err := DialWebSocket()
	println("Foo:", s, err)
}
