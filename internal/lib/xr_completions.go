package lib

import (
	"encoding/xml"
	"netconf-go/internal/types"
	"strings"

	"github.com/nemith/netconf"
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

var intfCache []string

// GetInterfaces @@@
func GetInterfaces(s *netconf.Session) []string {
	if intfCache != nil {
		return intfCache
	}
	// TODO make common with set
	// TODO make request more specific - just get names
	// requestLine := "get-oper openconfig-interfaces interfaces interface"
	requestLine := "get-oper openconfig-interfaces interfaces interface name= state name"
	slice := strings.Split(requestLine, " ")
	log.Debug("Set line:", slice[1:])

	// requestMap := make(map[string]interface{})
	// requestMap = expand(requestMap, slice[1:])
	// log.Debugf("expand: %v\n", requestMap)

	data, _ := SendNetconfRequest(s, requestLine, types.GetOper)
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
	// fmt.Printf("Intfs: %v\n", intfs)
	intfCache = intfs

	return intfs
}

type nodesReply struct {
	XMLName  xml.Name `xml:"data"`
	Text     string   `xml:",chardata"`
	Platform struct {
		Text  string `xml:",chardata"`
		Xmlns string `xml:"xmlns,attr"`
		Racks struct {
			Text string `xml:",chardata"`
			Rack []struct {
				Text     string `xml:",chardata"`
				RackName string `xml:"rack-name"`
				Slots    struct {
					Text string `xml:",chardata"`
					Slot []struct {
						Text      string `xml:",chardata"`
						SlotName  string `xml:"slot-name"`
						Instances struct {
							Text     string `xml:",chardata"`
							Instance []struct {
								Text         string `xml:",chardata"`
								InstanceName string `xml:"instance-name"`
								State        struct {
									Text                string `xml:",chardata"`
									CardType            string `xml:"card-type"`
									CardRedundancyState string `xml:"card-redundancy-state"`
									State               string `xml:"state"`
									AdminState          string `xml:"admin-state"`
									NodeName            string `xml:"node-name"`
									OperState           string `xml:"oper-state"`
								} `xml:"state"`
							} `xml:"instance"`
						} `xml:"instances"`
					} `xml:"slot"`
				} `xml:"slots"`
			} `xml:"rack"`
		} `xml:"racks"`
	} `xml:"platform"`
}

// GetNodes @@@
func GetNodes(s *netconf.Session) []string {
	// TODO make common with set
	requestLine := "get-oper Cisco-IOS-XR-platform-oper platform racks rack rack-name=0"

	data, _ := SendNetconfRequest(s, requestLine, types.GetOper)
	yangReply := nodesReply{}
	err := xml.Unmarshal([]byte(data), &yangReply)
	if err != nil {
		panic(err)
	}
	slots := make([]string, len(yangReply.Platform.Racks.Rack[0].Slots.Slot))
	var n int = 0
	for _, slot := range yangReply.Platform.Racks.Rack[0].Slots.Slot {
		if len(slot.Instances.Instance) > 0 {
			slots[n] = slot.Instances.Instance[0].State.NodeName
			n++
		}
	}
	// fmt.Printf("Nodes: %v\n", slots)

	return slots[:n]
}
