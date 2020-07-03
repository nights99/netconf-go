package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"

	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/chzyer/readline"
	"github.com/openconfig/goyang/pkg/yang"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Array of available Yang modules
var modNames []string

var mods = map[string]*yang.Module{}

var ms *yang.Modules

var globalSession *netconf.Session

const (
	validate = 0
	commit   = 1
	getConf  = 2
	getOper  = 3
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("set", readline.PcItemDynamic(listYang)),
	readline.PcItem("get-conf", readline.PcItemDynamic(listYang)),
	readline.PcItem("get-oper", readline.PcItemDynamic(listYang)),
	readline.PcItem("validate"),
	readline.PcItem("commit"))

type netconfRequest struct {
	ncEntry     yang.Entry
	NetConfPath []string
	Value       string
	reqType     int
}

type schemaJ struct {
	Identifier string `xml:"identifier"`
	//Version    string `xml:"version"`
	//Format     string `xml:"format"`
	//Namespace  string `xml:"namespace"`
	//Location    string  `xml:"location"`
}

type schemaReply3 struct {
	XMLName xml.Name
	Schemas []schemaJ `xml:"schema"`
}

type schemaReply2 struct {
	XMLName xml.Name
	Rest    schemaReply3 `xml:"schemas"`
}

type schemaReply struct {
	XMLName xml.Name     `xml:"data"`
	Rest    schemaReply2 `xml:"netconf-state"`
}

type yangReply struct {
	XMLName xml.Name `xml:"data"`
	Rest    string   `xml:",chardata"`
}

func expand(expandedMap map[string]interface{}, value []string) map[string]interface{} {
	log.Debugf("map: %v, value: %s\n", expandedMap, value)
	if len(value) == 1 {
		expandedMap[value[0]] = ""
	} else if len(value) == 2 {
		expandedMap[value[0]] = value[1]
	} else {
		if expandedMap[value[0]] == nil {
			expandedMap[value[0]] = make(map[string]interface{})
		}
		expandedMap[value[0]] = expand(expandedMap[value[0]].(map[string]interface{}), value[1:])
	}

	return expandedMap
}

func newNetconfRequest(netconfEntry yang.Entry, Path []string, value string, requestType int) *netconfRequest {
	return &netconfRequest{
		NetConfPath: Path,
		ncEntry:     netconfEntry,
		Value:       value,
		reqType:     requestType,
	}
}

func emitNestedXML(enc *xml.Encoder, paths []string, value string) {
	start3 := xml.StartElement{Name: xml.Name{Local: paths[0]}}
	err := enc.EncodeToken(start3)
	if err != nil {
		fmt.Println(err)
	}
	if len(paths) > 1 {
		emitNestedXML(enc, paths[1:], "")
	} else if value != "" {
		enc.EncodeToken(xml.CharData(value))
	}
	err = enc.EncodeToken(start3.End())
	if err != nil {
		fmt.Println(err)
	}
}

func (nc *netconfRequest) MarshalMethod() string {
	var buf bytes.Buffer

	enc := xml.NewEncoder(&buf)

	switch nc.reqType {
	case commit:
		fallthrough
	case validate:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "edit-config"}})
		emitNestedXML(enc, []string{"target", "candidate"}, "")
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "config", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}})
	case getConf:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "get-config"}})
		emitNestedXML(enc, []string{"source", "running"}, "")
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "filter"}})
	case getOper:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "get"}})
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "filter"}})
	}

	start2 := xml.StartElement{Name: xml.Name{Local: nc.NetConfPath[0], Space: nc.ncEntry.Namespace().Name}}
	//fmt.Println(start2)
	err := enc.EncodeToken(start2)
	if err != nil {
		fmt.Println(err)
	}

	emitNestedXML(enc, nc.NetConfPath[1:], nc.Value)

	err = enc.EncodeToken(start2.End())
	if err != nil {
		fmt.Println(err)
	}
	switch nc.reqType {
	case commit:
		fallthrough
	case validate:
		err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "config", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}})
		if err != nil {
			fmt.Println(err)
		}
		err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "edit-config"}})
		if err != nil {
			fmt.Println(err)
		}
	case getConf:
		enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "filter"}})
		enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "get-config"}})
	case getOper:
		enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "filter"}})
		enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "get"}})
	}

	enc.Flush()

	return buf.String()
}

func listYang(path string) []string {
	log.Debugf("listYang called on path: %s", path)
	names := make([]string, 0)
	/*files, _ := ioutil.ReadDir(path)
	  for _, f := range files {
	      names = append(names, f.Name())
	  }
	  return names*/

	tokens := strings.Fields(path)
	log.Debugf("tokens: %d, %v", len(tokens), tokens)

	if len(tokens) >= 2 {
		// We have a module name; check for partial or incorrect
		if i := sort.SearchStrings(modNames, tokens[1]); i == len(modNames) || modNames[i] != tokens[1] {
			return modNames
		}
		if mods[tokens[1]] == nil {
			mods[tokens[1]] = getYangModule(globalSession, tokens[1])
		}
		mod := mods[tokens[1]]
		if mod == nil {
			return modNames
		}
		// TODO move this into getYangModule
		for _, i := range mod.Include {
			log.Debugf("Mod: %v %v", mod.Name, i)
			submod := getYangModule(globalSession, i.Name)
			if submod == nil {
				log.Infof("Failed to get %v", i.Name)
			} else {
				yang.ToEntry(submod)
			}
		}
		ms.Modules[ms.Modules[tokens[1]].FullName()] = nil
		mods[tokens[1]] = getYangModule(globalSession, tokens[1])
		ms.Process()

		entry := yang.ToEntry(mod)
		for _, e := range tokens[2:] {
			if entry != nil && e != "" {
				log.Debugf("Foo: e %v kind %v %v\n", e, entry.Kind, entry)
				if entry.Kind == yang.DirectoryEntry {
					prevEntry := entry
					entry = entry.Dir[e]
					if entry == nil {
						log.Debugf("Couldn't find %v in %v", e, prevEntry.Dir)
						entry = prevEntry
						tokens = tokens[:len(tokens)-1]
					}
				}

			}
		}
		if entry != nil {
			log.Debugf("Entry: kind %v dir %v Uses: %v", entry.Kind, entry.Dir, entry.Errors)
		}
		if entry != nil && entry.Kind == yang.DirectoryEntry {
			for s := range entry.Dir {
				names = append(names, strings.Join(tokens[1:], " ")+" "+s)
			}
		}

		// 	if len(tokens) > 3 {
		// 		if len(tokens) > 4 {
		// 			entry := yang.ToEntry(mod).Dir[tokens[2]]
		// 			log.Debugf("Foo: %v\n", entry)
		// 		} else {
		// 			for s := range entry.Dir {
		// 				log.Debugf("Foo: %v\n", s)
		// 				names = append(names, tokens[1]+" "+tokens[2]+" "+s)
		// 			}
		// 		}
		// 	} else {
		// 		log.Debugf("Yang mod: type %T", mod)
		// 		entry := yang.ToEntry(mod)
		// 		log.Debugf("Yang mod: %v", entry.Kind)

		// 		if entry.Kind == yang.DirectoryEntry {
		// 			for s := range entry.Dir {
		// 				names = append(names, tokens[1]+" "+s)
		// 			}
		// 		}
		// 	}
		log.Debugf("names: %v\n", names)
	} else {
		log.Debug("Returning all modules")
		names = modNames
	}
	return names
}

func getSchemaList(s *netconf.Session) []string {
	/*
	 * Get a list of schemas
	 */
	reply, error := s.Exec(netconf.RawMethod(`<get>
    <filter type="subtree">
      <netconf-state xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring">
        <schemas/>
      </netconf-state>
    </filter>
    </get>`))
	if error != nil {
		fmt.Printf("Request reply error: %v\n", error)
	}
	//fmt.Printf("Request reply: %v, error: %v\n", reply.Data, error)
	schemaReply := schemaReply{}
	error = xml.Unmarshal([]byte(reply.Data), &schemaReply)
	//fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[0], err)
	//fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[99].Identifier, err)
	if error != nil {
		fmt.Printf("Request reply error: %v\n", error)
	}

	var schStrings []string
	for _, sch := range schemaReply.Rest.Rest.Schemas {
		schStrings = append(schStrings, sch.Identifier)
	}

	sort.Strings(schStrings)
	return schStrings
}

func getYangModule(s *netconf.Session, yangMod string) *yang.Module {
	/*
	 * Get the yang module from XR and read it into the map
	 */
	//log.Debug("Getting: ", yangMod)
	reply, error := s.Exec(netconf.RawMethod(`<get-schema
		 xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring">
	 <identifier>` +
		yangMod +
		`</identifier>
		 </get-schema>
	 `))
	if error != nil {
		fmt.Printf("Request reply error: %v\n", error)
		return nil
	}
	//log.Debugf("Request reply: %v, error: %v\n", reply, error)
	yangReply := yangReply{}
	err := xml.Unmarshal([]byte(reply.Data), &yangReply)
	//fmt.Printf("Request reply: %v, error: %v\n", yangReply, err)
	err = ms.Parse(yangReply.Rest, yangMod)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic("Error")
	}
	if ms.Modules[yangMod] != nil {
		return ms.Modules[yangMod]
	} else if ms.SubModules[yangMod] != nil {
		return ms.SubModules[yangMod]
	}
	return nil
}

func sendNetconfRequest(s *netconf.Session, requestLine string, requestType int) {
	slice := strings.Split(requestLine, " ")

	// Create a request structure with module, path array, and string value.
	var ncRequest *netconfRequest
	switch requestType {
	case commit:
		fallthrough
	case validate:
		ncRequest = newNetconfRequest(*yang.ToEntry(mods[slice[1]]), slice[2:len(slice)-1], slice[len(slice)-1], requestType)
	case getConf:
		fallthrough
	case getOper:
		ncRequest = newNetconfRequest(*yang.ToEntry(mods[slice[1]]), slice[2:], "", requestType)
	default:
		panic("Bad request type")
	}

	//fmt.Printf("ncRequest: %v\n", ncRequest)

	rpc := netconf.NewRPCMessage([]netconf.RPCMethod{ncRequest})
	xml2, err := xml.MarshalIndent(rpc, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	log.Debug(string(xml2))

	reply, error := s.Exec(ncRequest)

	log.Debugf("Request reply: %v, error: %v\n", reply, error)

	if requestType == commit {
		reply, error = s.Exec(netconf.RawMethod("<commit></commit>"))
		log.Debugf("Request reply: %v, error: %v\n", reply, error)
	} else if requestType == validate {
		reply, error = s.Exec(netconf.RawMethod("<validate><source><candidate/></source></validate>"))
		log.Debugf("Request reply: %v, error: %v\n", reply, error)
	} else if requestType == getConf || requestType == getOper {
		log.Debugf("Request reply: %v, error: %v, data: %v\n", reply, error, reply.Data)

		dec := xml.NewDecoder(strings.NewReader(reply.Data))
		var tok xml.Token
		var lastString string
		var theString string
		var seenFirstEnd bool
		seenFirstEnd = false
		for {
			tok, error = dec.Token()
			//fmt.Printf("Token: %T\n", tok, error)
			switch v := tok.(type) {
			case xml.CharData:
				//fmt.Printf("Token: %v\n", string(v), error)
				lastString = string(v)
			case xml.EndElement:
				if !seenFirstEnd {
					seenFirstEnd = true
					theString = lastString
				}

			default:
				//fmt.Printf("Token: %v\n", v, error)
			}
			if tok == nil {
				break
			}
		}
		fmt.Println("Data: ", theString)

	}
}

func main() {
	// Parse args
	var port = flag.Int("port", 10555, "Port number to connect to")
	var addr = flag.String("address", "localhost", "Address or host to connect to")
	var logLevel = flag.String("debug", log.InfoLevel.String(), "debug level")
	flag.Parse()

	l2, _ := log.ParseLevel(*logLevel)
	log.SetLevel(l2)

	// Connect to the node
	//s, err := netconf.DialTelnet("localhost:"+strconv.Itoa(*port), "lab", "lab", nil)

	//sshConfig, err := netconf.SSHConfigPubKeyFile("root", "/users/jnightin/.ssh/id_moonshine", "")
	// if err != nil {
	//     panic(err)
	// }
	sshConfig := netconf.SSHConfigPassword("cisco", "cisco123")
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	s, err := netconf.DialSSH(*addr+":"+strconv.Itoa(*port), sshConfig)

	if err != nil {
		panic(err)
	}
	globalSession = s

	defer s.Close()

	//fmt.Printf("Server Capabilities: '%+v'\n", s.ServerCapabilities)
	//fmt.Printf("Session Id: %d\n\n", s.SessionID)

	ms = yang.NewModules()

	realMods := true
	if realMods {
		modNames = getSchemaList(s)
		//fmt.Printf("modNames: %v\n", modNames)
	} else {
		if err := ms.Read("Cisco-IOS-XR-shellutil-cfg.yang"); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if err := ms.Read("Cisco-IOS-XR-cdp-cfg.yang"); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		//fmt.Printf("%v\n", ms)

		for _, m := range ms.Modules {
			if mods[m.Name] == nil {
				mods[m.Name] = m
				modNames = append(modNames, m.Name)
			}
		}
	}
	sort.Strings(modNames)
	//println(modNames)
	//fmt.Printf("names: %v\n", modNames)
	//entries := make([]*yang.Entry, len(modNames))
	//for x, n := range modNames {
	//	entries[x] = yang.ToEntry(mods[n])
	//}
	//fmt.Printf("+%v\n", entries[0])
	//for _, e := range entries {
	//	//print(e.Description)
	//	fmt.Printf("\n\n\n\n")
	//	//e.Print(os.Stdout)
	//	for s1, e1 := range e.Dir {
	//		println(s1)
	//		e1.Print(os.Stdout)
	//	}
	//}

	l, err := readline.NewEx(&readline.Config{
		Prompt:            "netconf> ",
		HistoryFile:       "/tmp/readline.tmp",
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		println("Error!")
		panic(err)
	}
	defer l.Close()
	//log.SetOutput(l.Stderr())
	var requestLine string

	for {
		// Maps string to void
		// Becomes a nested map of strings
		requestMap := make(map[string]interface{})
		//println("In loop")
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "set"):
			requestLine = line
			slice := strings.Split(requestLine, " ")
			log.Debug("Set line:", slice[1:])

			requestMap = expand(requestMap, slice[1:])

			log.Debugf("expand: %v\n", requestMap)

			/*
			 * If we don't know the module, read it from the router now.
			 */
			if mods[slice[1]] == nil {
				mods[slice[1]] = getYangModule(s, slice[1])
			}
			break
		case strings.HasPrefix(line, "get-conf"):
			// TODO make common with set
			requestLine = line
			slice := strings.Split(requestLine, " ")
			log.Debug("Set line:", slice[1:])

			requestMap = expand(requestMap, slice[1:])

			log.Debugf("expand: %v\n", requestMap)

			/*
			 * If we don't know the module, read it from the router now.
			 */
			if mods[slice[1]] == nil {
				mods[slice[1]] = getYangModule(s, slice[1])
				if mods[slice[1]] == nil {
					continue
				}
			}
			sendNetconfRequest(s, requestLine, getConf)
			break
		case strings.HasPrefix(line, "get-oper"):
			// TODO make common with set
			requestLine = line
			slice := strings.Split(requestLine, " ")
			log.Debug("Set line:", slice[1:])

			requestMap = expand(requestMap, slice[1:])

			log.Debugf("expand: %v\n", requestMap)

			/*
			 * If we don't know the module, read it from the router now.
			 */
			if mods[slice[1]] == nil {
				mods[slice[1]] = getYangModule(s, slice[1])
				if mods[slice[1]] == nil {
					continue
				}
			}
			sendNetconfRequest(s, requestLine, getOper)
			break
		case strings.HasPrefix(line, "validate"):
			sendNetconfRequest(s, requestLine, validate)
			break
		case strings.HasPrefix(line, "commit"):
			sendNetconfRequest(s, requestLine, commit)
			break
		default:
		}
		log.Debug("you said:", strconv.Quote(line))
	}
}
