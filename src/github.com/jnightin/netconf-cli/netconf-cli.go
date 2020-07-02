package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/chzyer/readline"
	"github.com/openconfig/goyang/pkg/yang"
	"golang.org/x/crypto/ssh"
)

// Array of available Yang modules
var mod_names []string

var mods = map[string]*yang.Module{}

var ms *yang.Modules

var global_session *netconf.Session


var completer = readline.NewPrefixCompleter(readline.PcItem("set", readline.PcItemDynamic(listYang)), readline.PcItem("validate"))

type netconfRequest struct {
	ncEntry     yang.Entry
	NetConfPath []string
	Value       string
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
	XMLName xml.Name     `xml:"data"`
	Rest    string		 `xml:",chardata"`
}

func Expand(expanded_map map[string]interface{}, value []string) map[string]interface{} {
	fmt.Printf("map: %v, value: %s\n", expanded_map, value)
	if len(value) == 2 {
		expanded_map[value[0]] = value[1]
	} else {
		if expanded_map[value[0]] == nil {
			expanded_map[value[0]] = make(map[string]interface{})
		}
		expanded_map[value[0]] = Expand(expanded_map[value[0]].(map[string]interface{}), value[1:])
	}

	return expanded_map
}


func newNetconfRequest(netconfEntry yang.Entry, Path []string, value string) *netconfRequest {
	return &netconfRequest{
		NetConfPath: Path,
		ncEntry:     netconfEntry,
		Value:       value,
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

	enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "edit-config"}})

	emitNestedXML(enc, []string{"target", "candidate"}, "")

	enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "config", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}})

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
	err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "config", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}})
	if err != nil {
		fmt.Println(err)
	}
	err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "edit-config"}})
	if err != nil {
		fmt.Println(err)
	}
	enc.Flush()

	return buf.String()
}

func listYang(path string) []string {
	fmt.Printf("\nlistYang called on path: %s\n", path)
	names := make([]string, 0)
	/*files, _ := ioutil.ReadDir(path)
	  for _, f := range files {
	      names = append(names, f.Name())
	  }
	  return names*/

	tokens := strings.Split(path, " ")
	//fmt.Printf("tokens: %v\n", tokens)

	if len(tokens) > 2 {
		// We have a module name
		if mods[tokens[1]] == nil {
			 mods[tokens[1]] = getYangModule(global_session, tokens[1])
		}
		mod := mods[tokens[1]]
		if len(tokens) > 3 {
			entry := yang.ToEntry(mod).Dir[tokens[2]]
            fmt.Printf("Foo: %v\n", entry)
            if len(tokens) > 4 {
                entry := yang.ToEntry(mod).Dir[tokens[2]]
                fmt.Printf("Foo: %v\n", entry)
            } else {
    			for s := range entry.Dir {
    				names = append(names, tokens[1]+" "+tokens[2]+" "+s)
    			}
            }
		} else {
			for s := range yang.ToEntry(mod).Dir {
				names = append(names, tokens[1]+" "+s)
			}
		}
		//fmt.Printf("names: %v\n", names)
	} else {
		//names = append(names, "hostname")
		//names = append(names, mod_names[0])
		names = mod_names
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
	err := xml.Unmarshal([]byte(reply.Data), &schemaReply)
	fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[0], err)
	fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[99].Identifier, err)

	var sch_strings []string
	for _, sch := range schemaReply.Rest.Rest.Schemas {
		sch_strings = append(sch_strings, sch.Identifier)
	}

	return sch_strings
}

func getYangModule (s *netconf.Session, yang_mod string) *yang.Module {
	/*
	 * Get the yang module from XR and read it into the map
	 */
	reply, error := s.Exec(netconf.RawMethod(`<get-schema
		 xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring">
	 <identifier>` +
		yang_mod +
		`</identifier>
		 </get-schema>
	 `))
	 if error != nil {
		 fmt.Printf("Request reply error: %v\n", error)
	 }
	 //fmt.Printf("Request reply: %v, error: %v\n", reply, error)
	 yangReply := yangReply{}
	 err := xml.Unmarshal([]byte(reply.Data), &yangReply)
	 //fmt.Printf("Request reply: %v, error: %v\n", yangReply, err)
	 err = ms.Parse(yangReply.Rest, yang_mod)
	 if err != nil {
		 fmt.Fprintln(os.Stderr, err)
	 }

	 return ms.Modules[yang_mod]
}

func sendNetconfRequest(s *netconf.Session, request_line string, commit bool) {
	slice := strings.Split(request_line, " ")

	ncRequest := newNetconfRequest(*yang.ToEntry(mods[slice[1]]), slice[2:len(slice)-1], slice[len(slice)-1])
	fmt.Printf("ncRequest: %v\n", ncRequest)

	rpc := netconf.NewRPCMessage([]netconf.RPCMethod{ncRequest})
	xml2, err := xml.MarshalIndent(rpc, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Stdout.Write(xml2)
	println()

	reply, error := s.Exec(ncRequest)

	fmt.Printf("Request reply: %v, error: %v\n", reply, error)

	println("Request sent!")

	if commit {
		reply, error = s.Exec(netconf.RawMethod("<commit></commit>"))
	} else {
		reply, error = s.Exec(netconf.RawMethod("<validate><source><candidate/></source></validate>"))
	}
	fmt.Printf("Request reply: %v, error: %v\n", reply, error)
}

func main() {

	// Parse args
	var port = flag.Int("port", 10555, "Port number to connect to")
	var addr = flag.String("address", "localhost", "Address or host to connect to")
	flag.Parse()

	// Connect to the node
	//s, err := netconf.DialTelnet("localhost:"+strconv.Itoa(*port), "lab", "lab", nil)


    //sshConfig, err := netconf.SSHConfigPubKeyFile("root", "/users/jnightin/.ssh/id_moonshine", "")
    // if err != nil {
    //     panic(err)
    // }
    sshConfig := netconf.SSHConfigPassword("cisco", "cisco123")
    sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
    s, err := netconf.DialSSH(*addr + ":"+strconv.Itoa(*port), sshConfig)


	if err != nil {
		panic(err)
	}
	global_session = s

	defer s.Close()

	//fmt.Printf("Server Capabilities: '%+v'\n", s.ServerCapabilities)
	fmt.Printf("Session Id: %d\n\n", s.SessionID)

	ms = yang.NewModules()

	real_mods := true
	if real_mods {
		mod_names = getSchemaList(s);
		//fmt.Printf("mod_names: %v\n", mod_names)
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
				mod_names = append(mod_names, m.Name)
			}
		}
	}
	sort.Strings(mod_names)
	//println(mod_names)
	//fmt.Printf("names: %v\n", mod_names)
	//entries := make([]*yang.Entry, len(mod_names))
	//for x, n := range mod_names {
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

	fmt.Printf("\n\n\n\n")

	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[31m»\033[0m ",
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
	log.SetOutput(l.Stderr())
	var request_line string

	request_map := make(map[string]interface{})

	for {
		println("In loop")
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
			request_line = line
			slice := strings.Split(request_line, " ")
			log.Println("Set line:", slice[1:])

			request_map = Expand(request_map, slice[1:])

			fmt.Printf("expand: %v\n", request_map)

			/*
			 * If we don't know the module, reads it from the router now.
			 */
			if mods[slice[1]] == nil {
				 mods[slice[1]] = getYangModule(s, slice[1])
			}
			break
		case strings.HasPrefix(line, "validate"):
			sendNetconfRequest(s, request_line, false)
			break
		case strings.HasPrefix(line, "commit"):
			sendNetconfRequest(s, request_line, true)
			break
		default:
		}
		log.Println("you said:", strconv.Quote(line))
	}
}
