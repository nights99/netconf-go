package main

import (
	"bytes"
	"encoding/xml"
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
)

var completer = readline.NewPrefixCompleter(readline.PcItem("get", readline.PcItemDynamic(listYang("./"))))

type netconfRequest struct {
	ncEntry     yang.Entry
	NetConfPath []string
	Value       string
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
	fmt.Println(start2)
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

func listYang(path string) func(string) []string {
	println("Outer func called")
	return func(line string) []string {
		names := make([]string, 0)
		/*files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names*/
		println(path)
		names = append(names, "hostname")
		return names
	}
}

func main() {

	println("Foo!")

	ms := yang.NewModules()

	if err := ms.Read("Cisco-IOS-XR-shellutil-cfg.yang"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Printf("%v\n", ms)

	mods := map[string]*yang.Module{}
	var names []string

	for _, m := range ms.Modules {
		if mods[m.Name] == nil {
			mods[m.Name] = m
			names = append(names, m.Name)
		}
	}
	sort.Strings(names)
	entries := make([]*yang.Entry, len(names))
	for x, n := range names {
		entries[x] = yang.ToEntry(mods[n])
	}
	fmt.Printf("+%v\n", entries[0])
	for _, e := range entries {
		//print(e.Description)
		fmt.Printf("\n\n\n\n")
		e.Print(os.Stdout)
		for s1, e1 := range e.Dir {
			println(s1)
			e1.Print(os.Stdout)
		}
	}

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
		case strings.HasPrefix(line, "get"):
		default:
			log.Println("you said:", strconv.Quote(line))
		}
	}

	var NetconfPath = "Cisco-IOS-XR-shellutil-cfg.host-names.host-name"

	println(netconf.MethodGetConfig(NetconfPath))
	//xml, err := xml.Marshal(map[int]string{1: "host-names", 2: "host-name"})
	//xml2, err := xml.Marshal(entries[0])
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	//os.Stdout.Write(xml2)
	//println()

	ncRequest := newNetconfRequest(*entries[0], []string{"host-names", "host-name"}, "CCV-invalid-hostname")
	rpc := netconf.NewRPCMessage([]netconf.RPCMethod{ncRequest})
	xml2, err := xml.Marshal(rpc)
	os.Stdout.Write(xml2)
	println()

	s, err := netconf.DialTelnet("localhost:34392", "lab", "lab", nil)
	if err != nil {
		panic(err)
	}

	defer s.Close()

	//fmt.Printf("Server Capabilities: '%+v'\n", s.ServerCapabilities)
	fmt.Printf("Session Id: %d\n\n", s.SessionID)
	s.Exec(ncRequest)
}
