package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/openconfig/goyang/pkg/yang"
	log "github.com/sirupsen/logrus"
)

const (
	validate = 0
	commit   = 1
	getConf  = 2
	getOper  = 3
)

type yangReply struct {
	XMLName xml.Name `xml:"data"`
	Rest    string   `xml:",chardata"`
}

type netconfPathElement struct {
	name  string
	value *string
}

type netconfRequest struct {
	ncEntry     yang.Entry
	NetConfPath []netconfPathElement
	Value       string
	reqType     int
}

type schemaReply struct {
	XMLName      xml.Name `xml:"data"`
	Text         string   `xml:",chardata"`
	NetconfState struct {
		Text    string `xml:",chardata"`
		Xmlns   string `xml:"xmlns,attr"`
		Schemas struct {
			Text   string `xml:",chardata"`
			Schema []struct {
				Text       string `xml:",chardata"`
				Identifier string `xml:"identifier"`
				Version    string `xml:"version"`
				Format     string `xml:"format"`
				Namespace  string `xml:"namespace"`
				Location   string `xml:"location"`
			} `xml:"schema"`
		} `xml:"schemas"`
	} `xml:"netconf-state"`
}

var ms *yang.Modules

var mods = map[string]*yang.Module{}

var globalSession *netconf.Session

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
						if entry.IsList() {
							// Assume this is a key value.
							// @@@ Check whether list key has been specified or not
							i := strings.Index(e, "=")
							fmt.Printf("Compare %v to %v, %d, %d\n", e, entry.Key, i, len(e))
							if i == -1 || i == len(e)-1 {
								tokens = tokens[:len(tokens)-1]
							}
						} else {
							tokens = tokens[:len(tokens)-1]
						}
					}
				}
				if entry.IsList() {
					log.Debugln("Found list: ", entry.Name, entry.Key)
				}

			}
		}
		if entry != nil {
			log.Debugf("Entry: kind %v dir %v Uses: %v", entry.Kind, entry.Dir, entry.Errors)
		}
		if entry.IsList() {
			fmt.Printf("Enter list key (%s, %s, %v)\n", entry.Key, entry.Dir[entry.Key].Description, entry.Dir[entry.Key].Type.Name)
			fmt.Printf("list key tokens: %v\n", tokens)
			e := tokens[len(tokens)-1]
			i := strings.Index(e, "=")
			fmt.Printf("Compare %v to %v, %d, %d\n", e, entry.Key, i, len(e))
			// if i == -1 || i == len(e)-1 {
			if i == -1 {
				if entry.Dir[entry.Key].Type.Name == "Interface-name" {
					intfs := GetInterfaces(globalSession)
					println(intfs)
					for _, intf := range intfs {
						names = append(names, strings.Join(tokens[1:], " ")+" "+entry.Key+"="+intf)
					}
				} else if entry.Dir[entry.Key].Type.Name == "Node-id" {
					nodes := GetNodes(globalSession)
					println(nodes)
					for _, node := range nodes {
						names = append(names, strings.Join(tokens[1:], " ")+" "+entry.Key+"="+node)
					}
				} else {
					names = append(names, strings.Join(tokens[1:], " ")+" "+entry.Key+"=")
				}
			} else {
				for s := range entry.Dir {
					names = append(names, strings.Join(tokens[1:], " ")+" "+s)
				}
			}
		} else if entry != nil && entry.Kind == yang.DirectoryEntry {
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

func newNetconfRequest(netconfEntry yang.Entry, Path []string, value string, requestType int) *netconfRequest {
	ncArray := make([]netconfPathElement, len(Path))
	for i, p := range Path {
		if strings.Contains(p, "=") {
			values := strings.Split(p, "=")
			ncArray[i].name = values[0]
			ncArray[i].value = &values[1]

		} else {
			ncArray[i].name = p
			ncArray[i].value = nil
		}
	}
	return &netconfRequest{
		NetConfPath: ncArray,
		ncEntry:     netconfEntry,
		Value:       value,
		reqType:     requestType,
	}
}

func emitNestedXML(enc *xml.Encoder, paths []netconfPathElement, value string) {
	start3 := xml.StartElement{Name: xml.Name{Local: paths[0].name}}
	err := enc.EncodeToken(start3)
	if err != nil {
		fmt.Println(err)
	}
	if paths[0].value != nil {
		enc.EncodeToken(xml.CharData(*paths[0].value))
		enc.EncodeToken(start3.End())
	}
	if len(paths) > 1 {
		emitNestedXML(enc, paths[1:], "")
	} else if value != "" {
		enc.EncodeToken(xml.CharData(value))
	}
	if paths[0].value == nil {
		err = enc.EncodeToken(start3.End())
		if err != nil {
			fmt.Println(err)
		}
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
		emitNestedXML(enc, []netconfPathElement{
			{name: "target", value: nil},
			{name: "candidate", value: nil}},
			"")
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "config", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}})
	case getConf:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "get-config"}})
		emitNestedXML(enc, []netconfPathElement{
			{name: "source", value: nil},
			{name: "running", value: nil}},
			"")
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "filter"}})
	case getOper:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "get"}})
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "filter"}, Attr: []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "subtree"}}})
	}

	start2 := xml.StartElement{Name: xml.Name{Local: nc.NetConfPath[0].name, Space: nc.ncEntry.Namespace().Name}}
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

func getYangModule(s *netconf.Session, yangMod string) *yang.Module {
	/*
	 * Get the yang module from XR and read it into the map
	 */
	log.Debug("Getting: ", yangMod)
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
	log.Debugf("Request reply: %v, error: %v\n", reply.Data, error)
	re, _ := regexp.Compile("\n#[0-9]+\n")
	// strs := re.FindAllStringSubmatch(reply.Data, 10)
	// fmt.Printf("%v\n", strs)
	reply.Data = re.ReplaceAllLiteralString(reply.Data, "")
	yangReply := yangReply{}
	err := xml.Unmarshal([]byte(reply.Data), &yangReply)
	//fmt.Printf("Request reply: %v, error: %v\n", yangReply, err)
	err = ms.Parse(yangReply.Rest, yangMod)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic("Error")
	}

	// TODO move this into getYangModule
	var mod *yang.Module = nil
	if ms.Modules[yangMod] != nil {
		mod = ms.Modules[yangMod]
	} else if ms.SubModules[yangMod] != nil {
		mod = ms.SubModules[yangMod]
	}
	if mod != nil {
		for _, i := range mod.Include {
			log.Debugf("Mod: %v %v", mod.Name, i)
			// Add check here whether we already have the submodule; if not get it, and note we need to reprocess this module further down.
			if ms.Modules[i.Name] == nil && ms.SubModules[i.Name] == nil {
				submod := getYangModule(globalSession, i.Name)
				if submod == nil {
					log.Infof("Failed to get %v", i.Name)
				} else {
					yang.ToEntry(submod)
				}
			} else {
				log.Debug("Already processed: ", i.Name)
			}
		}
	}
	if mod != nil {
		for _, i := range mod.Import {
			log.Debugf("Mod: %v %v", mod.Name, i)
			// Add check here whether we already have the submodule; if not get it, and note we need to reprocess this module further down.
			if ms.Modules[i.Name] == nil && ms.SubModules[i.Name] == nil {
				submod := getYangModule(globalSession, i.Name)
				if submod == nil {
					log.Infof("Failed to get %v", i.Name)
				} else {
					yang.ToEntry(submod)
				}
			} else {
				log.Debug("Already processed: ", i.Name)
			}
		}
	}
	if ms.Modules[yangMod] != nil {
		ms.Modules[ms.Modules[yangMod].FullName()] = nil
	} else if ms.SubModules[yangMod] != nil {
		ms.SubModules[ms.SubModules[yangMod].FullName()] = nil
	}
	//mods[yangMod] = getYangModule(globalSession, yangMod)
	err = ms.Parse(yangReply.Rest, yangMod)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic("Error")
	}
	if ms.Modules[yangMod] != nil {
		mod = ms.Modules[yangMod]
	} else if ms.SubModules[yangMod] != nil {
		mod = ms.SubModules[yangMod]
	}
	mods[yangMod] = mod

	//if ms.Modules[yangMod] != nil {
	log.Debugf("About to process %s", mod.Name)
	//yang.ToEntry(mod)
	ms.Process()
	log.Debugf("Stored and re-processed %s", mod.Name)
	//}

	return mod
}
func sendNetconfRequest(s *netconf.Session, requestLine string, requestType int) string {
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

	//log.Debugf("Request reply: %v, error: %v\n", reply, error)

	if requestType == commit {
		reply, error = s.Exec(netconf.RawMethod("<commit></commit>"))
		log.Debugf("Request reply: %v, error: %v\n", reply, error)
	} else if requestType == validate {
		reply, error = s.Exec(netconf.RawMethod("<validate><source><candidate/></source></validate>"))
		log.Debugf("Request reply: %v, error: %v\n", reply, error)
	} else if requestType == getConf || requestType == getOper {
		if error != nil {
			fmt.Printf("Request reply: %v, error: %v\n", reply, error)
			return ""
		}
		log.Debugf("Request reply: %v, error: %v, data: %v\n", reply, error, reply.Data)
		// fmt.Printf("Request data: %v\n", reply.Data)

		dec := xml.NewDecoder(strings.NewReader(reply.Data))
		var tok xml.Token
		var lastString string
		var theString string
		var seenFirstEnd bool
		seenFirstEnd = false
		for {
			tok, error = dec.Token()
			// fmt.Printf("Token: %T %v\n", tok, error)
			switch v := tok.(type) {
			case xml.CharData:
				// fmt.Printf("Token: %v %v\n", string(v), error)
				lastString = string(v)
				// theString = lastString
			case xml.EndElement:
				if !seenFirstEnd {
					seenFirstEnd = true
					theString = lastString
				}

			default:
				// fmt.Printf("Token: %v %v\n", v, error)
			}
			if tok == nil {
				break
			}
		}
		fmt.Println("Data: ", theString)

	}
	return reply.Data
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
	// fmt.Printf("Request reply: %v, error: %v\n", reply.Data[0:1000], error)
	schemaReply := schemaReply{}
	error = xml.Unmarshal([]byte(reply.Data), &schemaReply)
	//fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[0], err)
	//fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[99].Identifier, err)
	if error != nil {
		fmt.Printf("Request reply error: %v\n", error)
	}

	var schStrings []string
	// for _, sch := range schemaReply.Rest.Rest.Schemas {
	for _, sch := range schemaReply.NetconfState.Schemas.Schema {
		schStrings = append(schStrings, sch.Identifier)
	}

	sort.Strings(schStrings)
	return schStrings
}
