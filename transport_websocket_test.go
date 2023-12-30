//go:build exclude

package main

import (
	"fmt"
	"strings"
	"testing"

	netconf "github.com/nemith/netconf"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/stretchr/testify/assert"
)

// Run with go test -v -count=1 transport_websocket_test.go netconf_lib.go xr_completions.go transport_websocket.go

var modNames []string

func TestWS(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	ms = yang.NewModules()
	var s *netconf.Session
	var err error
	assert.NotPanics(t, func() { s, err = DialWebSocket("localhost", 12345) })
	println("Foo:", s, err)
	t.Log("Foo:", s, err)
	globalSession = s
	modNames = getSchemaList(s)
	fmt.Printf("Modname: %v\n", modNames[:3])

	globalSession = s
	entries, _ := listYang("get-oper Cisco-IOS-XR-shellutil-oper")
	fmt.Printf("listYang returned %v\n", entries)

	slice := strings.Split("Cisco-IOS-XR-shellutil-oper", " ")
	webEntries := make([]string, 0)
	webEntries = append(webEntries, "GetEntries")
	webEntries = append(webEntries, slice...)
	webEntries = append(webEntries, ":")

	// Now we need each entry at this directory level.
	for _, v := range entries {
		fmt.Printf("listYang returned %v\n", v)
		x := strings.Split(v, " ")
		webEntries = append(webEntries, x[len(x)-1:]...)
	}
	fmt.Printf("webEntries: %v\n", strings.Join(webEntries, " "))
}
