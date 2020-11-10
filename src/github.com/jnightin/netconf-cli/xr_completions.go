package main

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	log "github.com/sirupsen/logrus"
)

type intfReply struct {
	XMLName    xml.Name `xml:"data"`
	Text       string   `xml:",chardata"`
	Interfaces struct {
		Text      string `xml:",chardata"`
		Xmlns     string `xml:"xmlns,attr"`
		Interface []struct {
			Text   string `xml:",chardata"`
			Name   string `xml:"name"`
			Config struct {
				Text string `xml:",chardata"`
				Name string `xml:"name"`
			} `xml:"config"`
		} `xml:"interface"`
	} `xml:"interfaces"`
}

// GetInterfaces @@@
func GetInterfaces(s *netconf.Session) []string {
	// TODO make common with set
	requestLine := "get-oper openconfig-interfaces interfaces interface"
	slice := strings.Split(requestLine, " ")
	log.Debug("Set line:", slice[1:])

	requestMap := make(map[string]interface{})
	requestMap = expand(requestMap, slice[1:])

	log.Debugf("expand: %v\n", requestMap)

	/*
	* If we don't know the module, read it from the router now.
	 */
	if mods[slice[1]] == nil {
		mods[slice[1]] = getYangModule(s, slice[1])
	}
	data := sendNetconfRequest(s, requestLine, getOper)
	yangReply := intfReply{}
	err := xml.Unmarshal([]byte(data), &yangReply)
	if err != nil {
		panic(err)
	}
	intfs := make([]string, len(yangReply.Interfaces.Interface))
	for i, intf := range yangReply.Interfaces.Interface {
		// intfs = append(intfs, i.Name)
		intfs[i] = intf.Name
	}
	fmt.Printf("Intfs: %v\n", intfs)

	return intfs
}
