package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	netconf "github.com/nemith/netconf"
	"github.com/openconfig/goyang/pkg/yang"
	log "github.com/sirupsen/logrus"
)

const (
	validate = 0
	commit   = 1
	getConf  = 2
	getOper  = 3
	rpcOp    = 5
)

const (
	newTokens   = 0
	replaceLast = 1
)

type yangReply struct {
	XMLName xml.Name `xml:"data"`
	Rest    string   `xml:",chardata"`
}

type netconfPathElement struct {
	name   string
	value  *string
	delete bool
}

type netconfRequest struct {
	ncEntry     *yang.Entry
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

func listYang(path string) ([]string, int) {
	var didAugment bool = false
	var currentPrefix string
	var returnType = newTokens

	log.Debugf("listYang called on path: %s", path)
	// yang.ParseOptions.IgnoreSubmoduleCircularDependencies = true
	names := make([]string, 0)

	tokens := strings.Fields(path)
	log.Debugf("tokens: %d, %v", len(tokens), tokens)

	if len(tokens) >= 2 {
		// We have a module name; check for partial or incorrect
		if i := sort.SearchStrings(modNames, tokens[1]); i == len(modNames) || modNames[i] != tokens[1] {
			log.Debugf("didn't find %s in %v, returning all, 1", tokens[1], len(modNames))
			return modNames, returnType
		}
		if mods[tokens[1]] == nil {
			mods[tokens[1]] = getYangModule(globalSession, tokens[1])
		}
		mod := mods[tokens[1]]
		if mod == nil {
			log.Debugf("didn't find %s in %v, returning all, 2", tokens[1], len(mods))
			return modNames, returnType
		}

		entry := yang.ToEntry(mod)
		currentPrefix = mod.Namespace.Name
		// println("currentPrefix", currentPrefix)
		var deletedLastToken bool = false
		for _, e := range tokens[2:] {
			if entry != nil && e != "" {
				log.Debugf("Foo: e %v kind %v %v\n", e, entry.Kind, entry)
				if entry.Kind == yang.DirectoryEntry {
					prevEntry := entry

					// Check if a module prefix has been specified and if so strip it.
					if strings.Contains(e, "@") {
						e = strings.Split(e, "@")[1]
					}

					entry = entry.Dir[e]
					if entry == nil && prevEntry.RPC != nil {
						entry = prevEntry.RPC.Input.Dir[e]
					}
					if entry == nil {
						log.Debugf("Couldn't find %v in %v", e, prevEntry.Dir)
						entry = prevEntry
						if entry.IsList() {
							// Assume this is a key value.
							// @@@ Check whether list key has been specified or not
							i := strings.Index(e, "=")
							// fmt.Printf("Compare %v to %v, %d, %d\n", e, entry.Key, i, len(e))
							// @@@ This is horrible, needs fixing, and messes up the web code.
							// Doesn't have an equals - assume this is a partial string, ignore it an d return the whole dir
							//  OR
							// = is last character of token OR
							// We're on the last token and we don't have a space - not ready to advance yet - what if we remove this case?
							if i == -1 ||
								(i == len(e)-1 && !strings.HasSuffix(path, " ") && e == tokens[len(tokens)-1]) ||
								(e == tokens[len(tokens)-1] && !strings.HasSuffix(path, " ")) {
								tokens = tokens[:len(tokens)-1]
								deletedLastToken = true
							}
						} else {
							// Check for augment type path
							// - get prefix - bit before colon
							// - look up prefix in imports - gives you the module being augmented
							if strings.Contains(e, ":") {
								aug_tokens := strings.FieldsFunc(e,
									func(r rune) bool {
										return r == ':' || r == '/'
									},
								)
								// For now, assume the augment is for the same
								// module; check this explicitly to at least
								// make this assumption obvious.
								for i := 2; i < len(aug_tokens); i += 2 {
									if !strings.HasPrefix(aug_tokens[0], aug_tokens[i]) {
										// TODO do something better than panic here
										panic("Augments with different prefixes not currently support: " + e)
									}
								}
								// println("Possible augment prefix:", aug_tokens[0])
								var augMod *yang.Import
								for _, augMod = range mod.Import {
									if augMod.Prefix.Name == aug_tokens[0] {
										// println("Found aug module:", augMod.Name)
										break
									}
								}
								if augMod.Prefix.Name == aug_tokens[0] {
									// println(augMod.Name, aug_tokens[1], aug_tokens[3])
									m2 := yang.ToEntry(augMod.Module)

									// Find entry in the augmented module, and
									// set the tokens to point there.
									entry = m2
									tokens = []string{tokens[0], augMod.Name}
									for i := 1; i < len(aug_tokens); i += 2 {
										entry = entry.Dir[aug_tokens[i]]
										tokens = append(tokens, aug_tokens[i])
									}
									didAugment = true
								}
							} else {
								tokens = tokens[:len(tokens)-1]
							}
						}
					}
				}
				if entry.IsList() {
					log.Debugln("Found list: ", entry.Name, entry.Key)
				}

			}
		}
		if entry != nil {
			log.Debugf("Entry: %v kind %v dir %v Errors: %v Augments: %v Augmented-by: %v Uses: %v", entry.Name, entry.Kind, entry.Dir, entry.Errors, entry.Augmented, entry.Augments, entry.Uses)
			if entry.Prefix != nil {
				// Need to store the prefix somewhere and add it when constructing the request.
				var prefix_ns string
				switch entry.Prefix.Parent.(type) {
				case *yang.Module:
					log.Debugln("Found prefix: ", entry.Prefix.Parent.(*yang.Module).Namespace.Name)
					// println("Found prefix: ", entry.Prefix.Parent.(*yang.Module).Namespace.Name)
					prefix_ns = entry.Prefix.Parent.(*yang.Module).Namespace.Name
				case *yang.BelongsTo:
					log.Debugln("Found prefix2: ", entry.Prefix.Parent.(*yang.BelongsTo).Name)
					// println("Found prefix2: ", entry.Prefix.Parent.(*yang.BelongsTo).Name)
					prefix_ns = entry.Prefix.Parent.(*yang.BelongsTo).Name
				}
				if currentPrefix != prefix_ns && !didAugment {
					currentPrefix = prefix_ns
					println("Changed prefix:", currentPrefix)
					// Add prefix to current i.e. last token, if it doesn't
					// already have one.
					// @@@ Should check prefix matches?
					if !strings.Contains(tokens[len(tokens)-1], "@") {
						tokens[len(tokens)-1] = currentPrefix + "@" + tokens[len(tokens)-1]
						didAugment = true
					}
				}

			}
		}
		var prefix string
		if !didAugment {
			prefix = ""
		} else {
			prefix = strings.Join(tokens[1:], " ") + " "
		}
		if entry.IsList() {
			// TODO Need to support multiple keys properly here
			var keys []string
			if entry.Key != "" {
				keys = strings.Split(entry.Key, " ")
				fmt.Printf("Enter list key (%s, %s, %v)\n", keys[0], entry.Dir[keys[0]].Description, entry.Dir[keys[0]].Type.Name)
			}
			// fmt.Printf("list key tokens: %v\n", tokens)
			e := tokens[len(tokens)-1]
			i := strings.Index(e, "=")
			// fmt.Printf("Compare %v to %v, %d, %d\n", e, entry.Key, i, len(e))
			// if i == -1 || i == len(e)-1 {
			if (i == -1 || (!deletedLastToken && !strings.HasSuffix(path, " "))) &&
				entry.Key != "" {
				if entry.Dir[keys[0]].Type.Name == "Interface-name" ||
					(entry.Name == "interface" && keys[0] == "name") {
					intfs := GetInterfaces(globalSession)
					println(intfs)
					for _, intf := range intfs {
						names = append(names, prefix+keys[0]+"="+intf)
					}
				} else if entry.Dir[keys[0]].Type.Name == "Node-id" {
					nodes := GetNodes(globalSession)
					println(nodes)
					for _, node := range nodes {
						names = append(names, prefix+keys[0]+"="+node)
					}
				} else {
					names = append(names, prefix+keys[0]+"=")
				}
			} else {
				for s := range entry.Dir {
					names = append(names, s)
				}
			}
		} else if entry != nil && entry.RPC != nil {
			for s := range entry.RPC.Input.Dir {
				names = append(names, strings.Join(tokens[1:], " ")+" "+s+"=")
			}
		} else if entry != nil && entry.Kind == yang.DirectoryEntry {
			for s := range entry.Dir {
				names = append(names, prefix+s)
			}
		} else if entry != nil && entry.Parent.Kind == yang.InputEntry {
			// This is a leaf specifying the input for the RPC, prompt the user
			// for input
			fmt.Printf("Enter RPC input: %s\n", entry.Description)
			names = append(names, entry.Name+"=")
			returnType = replaceLast
		}
		for _, s := range mod.Augment {
			log.Debugln("Mod augment: ", s.Name)
			if !didAugment {
				// This isn't quite right.
				// names = append(names, strings.Join(tokens[1:], " ")+" "+s.Name)
				names = append(names, s.Name)
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
	return names, returnType
}

func newNetconfRequest(netconfEntry *yang.Entry, Path []string, value string, requestType int, delete bool) *netconfRequest {
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
		if i == len(Path)-1 {
			ncArray[i].delete = delete
		}
	}
	return &netconfRequest{
		NetConfPath: ncArray,
		ncEntry:     netconfEntry,
		Value:       value,
		reqType:     requestType,
	}
}

func emitNestedXML(enc *xml.Encoder, paths []netconfPathElement, value string, reqType int) {
	var start3 xml.StartElement
	if paths[0].delete {
		start3 = xml.StartElement{
			Name: xml.Name{Local: paths[0].name},
			// Possible operations are merge, replace, create, delete, remove
			// Use remove rather than delete so as not to error if config doesn't exist.
			Attr: []xml.Attr{{Name: xml.Name{Local: "operation", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}, Value: "remove"}}}
	} else {
		if strings.Contains(paths[0].name, "@") {
			elems := strings.Split(paths[0].name, "@")
			start3 = xml.StartElement{Name: xml.Name{Space: elems[0], Local: elems[1]}}
		} else {
			start3 = xml.StartElement{Name: xml.Name{Local: paths[0].name}}
		}

	}
	err := enc.EncodeToken(start3)
	if err != nil {
		fmt.Println(err)
	}
	if paths[0].value != nil {
		enc.EncodeToken(xml.CharData(*paths[0].value))
		enc.EncodeToken(start3.End())
	}
	if len(paths) > 1 {
		emitNestedXML(enc, paths[1:], value, reqType)
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

func (nc *netconfRequest) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {

	switch nc.reqType {
	case commit:
		fallthrough
	case validate:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "edit-config"}})
		emitNestedXML(enc, []netconfPathElement{
			{name: "target", value: nil},
			{name: "candidate", value: nil}},
			"",
			nc.reqType)
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "config", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}})

	case getConf:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "get-config"}})
		emitNestedXML(enc, []netconfPathElement{
			{name: "source", value: nil},
			// TODO - add an option to choose between views.
			{name: "running", value: nil}},
			// {name: "running-inheritance", value: nil}},
			"", nc.reqType)
		if nc.ncEntry != nil {
			enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "filter"}})
		}
	case getOper:
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "get"}})
		enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "filter"}, Attr: []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "subtree"}}})
	case rpcOp:
		// enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "rpc"}})
	}

	var err error
	if nc.ncEntry != nil {
		start2 := xml.StartElement{Name: xml.Name{Local: nc.NetConfPath[0].name, Space: nc.ncEntry.Namespace().Name}}
		//fmt.Println(start2)
		err = enc.EncodeToken(start2)
		if err != nil {
			fmt.Println(err)
		}

		if len(nc.NetConfPath) > 1 {
			emitNestedXML(enc, nc.NetConfPath[1:], nc.Value, nc.reqType)
		}

		err = enc.EncodeToken(start2.End())
		if err != nil {
			fmt.Println(err)
		}
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
		if nc.ncEntry != nil {
			enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "filter"}})
		}
		enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "get-config"}})
	case getOper:
		enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "filter"}})
		enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "get"}})
	case rpcOp:
		// enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "rpc"}})
	}

	enc.Flush()

	return nil
}

func getYangModule(s *netconf.Session, yangMod string) *yang.Module {
	/*
	 * Get the yang module from XR and read it into the map
	 */
	log.Debug("Getting: ", yangMod)
	reply, error := s.Do(context.Background(),
		`<get-schema
		 xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring">
	 <identifier>`+
			yangMod+
			`</identifier>
		 </get-schema>
	 `)
	if error != nil {
		fmt.Printf("Request reply error1: %v\n", error)
		return nil
	}
	// log.Debugf("Request reply: %v, error: %v\n", reply.Data, error)
	// re, _ := regexp.Compile("\n#[0-9]+\n")
	// strs := re.FindAllStringSubmatch(reply.Data, 10)
	// fmt.Printf("%v\n", strs)
	// reply.Data = re.ReplaceAllLiteralString(reply.Data, "")
	yangReply := yangReply{}
	_ = xml.Unmarshal([]byte(reply.Body), &yangReply)
	//fmt.Printf("Request reply: %v, error: %v\n", yangReply, err)
	err := ms.Parse(yangReply.Rest, yangMod)
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
			log.Debugf("Mod include: %v %v", mod.Name, i)
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
			log.Debugf("Mod import: %v %v", mod.Name, i)
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
func sendNetconfRequest(s *netconf.Session, requestLine string, requestType int) (string, string) {
	defer timeTrack(time.Now(), "Request")

	slice := strings.Split(requestLine, " ")

	// Create a request structure with module, path array, and string value.
	var ncRequest *netconfRequest
	switch requestType {
	case commit:
		fallthrough
	case validate:
		if slice[0] == "delete" {
			ncRequest = newNetconfRequest(yang.ToEntry(mods[slice[1]]), slice[2:], "", requestType, true)
		} else {
			ncRequest = newNetconfRequest(yang.ToEntry(mods[slice[1]]), slice[2:len(slice)-1], slice[len(slice)-1], requestType, false)
		}
	case getOper, getConf, rpcOp:
		if len(slice) >= 3 {
			ncRequest = newNetconfRequest(yang.ToEntry(mods[slice[1]]), slice[2:], "", requestType, false)
		} else if requestType == getConf {
			// getConf supports getting the whole config.
			ncRequest = newNetconfRequest(nil, []string{}, "", requestType, false)
		}
	default:
		panic("Bad request type")
	}

	// fmt.Printf("ncRequest: %v\n", ncRequest)

	rpc := ncRequest
	// xml2, err := xml.MarshalIndent(ncRequest, "", "  ")
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }
	// log.Debug(string(xml2))

	reply, error := s.Do(context.Background(), &rpc)

	// log.Debugf("Request reply: %s, error: %v\n", reply, error)
	var theString string

	if requestType == commit {
		error = s.Commit(context.Background())
		log.Debugf("Request reply: %v, error: %v\n", reply, error)
	} else if requestType == validate {
		error = s.Validate(context.Background(), netconf.Candidate)
		log.Debugf("Request reply: %v, error: %v\n", reply, error)
	} else if requestType == getConf || requestType == getOper {
		if error != nil {
			fmt.Printf("Request reply: %v, error: %v\n", reply, error)
			return "", ""
		}
		log.Debugf("Request reply: %v, error: %v, data: %v\n", reply, error, reply.Body)
		// fmt.Printf("Request data: %v\n", reply.Data)

		dec := xml.NewDecoder(bytes.NewReader(reply.Body))
		var tok xml.Token
		var lastString string
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
		// TODO Handle bool/presence type items
		fmt.Println("Data: ", theString)

	}
	if reply != nil {
		return string(reply.Body), theString
	} else {
		return error.Error(), ""
	}
}

func getSchemaList(s *netconf.Session) []string {
	/*
	 * Get a list of schemas
	 */
	reply, error := s.Do(context.Background(), `<get>
    <filter type="subtree">
      <netconf-state xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring">
        <schemas/>
      </netconf-state>
    </filter>
    </get>`)
	if error != nil {
		fmt.Printf("Request reply error2: %v\n", error)
		// panic(error)
		return nil
	}
	// fmt.Printf("Request reply: %v, error: %v\n", reply.Data[0:1000], error)
	schemaReply := schemaReply{}
	error = xml.Unmarshal([]byte(reply.Body), &schemaReply)
	//fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[0], err)
	//fmt.Printf("Request reply: %v, error: %v\n", schemaReply.Rest.Rest.Schemas[99].Identifier, err)
	if error != nil {
		fmt.Printf("Request reply error3: %v\n", error)
	}

	var schStrings []string
	// for _, sch := range schemaReply.Rest.Rest.Schemas {
	for _, sch := range schemaReply.NetconfState.Schemas.Schema {
		schStrings = append(schStrings, sch.Identifier)
	}

	sort.Strings(schStrings)
	return schStrings
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
