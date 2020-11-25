// +build !wasm

package main

import (
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

var completer = readline.NewPrefixCompleter(
	readline.PcItem("set", readline.PcItemDynamic(listYang)),
	readline.PcItem("get-conf", readline.PcItemDynamic(listYang)),
	readline.PcItem("get-oper", readline.PcItemDynamic(listYang)),
	readline.PcItem("validate"),
	readline.PcItem("commit"))

// type schemaJ struct {
// 	Identifier string `xml:"identifier"`
// 	//Version    string `xml:"version"`
// 	//Format     string `xml:"format"`
// 	//Namespace  string `xml:"namespace"`
// 	//Location    string  `xml:"location"`
// }

// func expand(expandedMap map[string]interface{}, value []string) map[string]interface{} {
// 	log.Debugf("map: %v, value: %s\n", expandedMap, value)
// 	if len(value) == 1 {
// 		expandedMap[value[0]] = ""
// 	} else if len(value) == 2 {
// 		expandedMap[value[0]] = value[1]
// 	} else {
// 		if expandedMap[value[0]] == nil {
// 			expandedMap[value[0]] = make(map[string]interface{})
// 		}
// 		expandedMap[value[0]] = expand(expandedMap[value[0]].(map[string]interface{}), value[1:])
// 	}

// 	return expandedMap
// }

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

	// fmt.Printf("Server Capabilities: '%+v'\n", s.ServerCapabilities[0])
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
	log.SetOutput(l.Stderr())
	var requestLine string

	for {
		// Maps string to void
		// Becomes a nested map of strings
		// requestMap := make(map[string]interface{})
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

			// requestMap = expand(requestMap, slice[1:])
			// log.Debugf("expand: %v\n", requestMap)

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

			// requestMap = expand(requestMap, slice[1:])
			// log.Debugf("expand: %v\n", requestMap)

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

			if len(slice) < 3 {
				continue
			}

			// requestMap = expand(requestMap, slice[1:])
			// log.Debugf("expand: %v\n", requestMap)

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
