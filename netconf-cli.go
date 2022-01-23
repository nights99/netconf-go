//go:build !wasm
// +build !wasm

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/chzyer/readline"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/peterh/liner"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Array of available Yang modules
var modNames []string

var historyFile = filepath.Join(os.TempDir(), ".liner_example_history")

func wordCompleter(line string, pos2 int) (head string, completions []string, tail string) {
	// fmt.Printf("Called with %v\n", line)

	// TODO check if pos is end of line

	tokens := strings.Fields(line)
	log.Debugf("tokens: %d, %v", len(tokens), tokens)

	if len(tokens) >= 2 || strings.HasSuffix(line, " ") {
		yangCompletions := listYang(line)
		// fmt.Printf("Completions %v\n", yangCompletions)

		cs := make([]string, len(yangCompletions))
		var pos int = 0
		for _, e := range yangCompletions {
			// fmt.Printf("Comparing '%s' and '%s'\n", e[strings.LastIndex(e, " ")+1:], tokens[len(tokens)-1])
			// if strings.HasPrefix(e[strings.LastIndex(e, " ")+1:], tokens[len(tokens)-1]) || strings.HasSuffix(line, " ") {
			if strings.HasPrefix(e, tokens[len(tokens)-1]) || strings.HasSuffix(line, " ") {
				// println("Found " + e)
				// cs[pos] = tokens[0] + " " + e
				cs[pos] = e
				pos++
			}
		}
		// cs = []string{longestcommon.Prefix(cs[:pos])}
		// fmt.Printf("Found %v\n", cs)
		// TODO Should we add a space on the end if we've found a completion?
		if len(cs[:pos]) == 1 {
			cs[0] += " "
		}
		if strings.HasSuffix(line, " ") {
			return strings.Join(tokens, " ") + " ", cs[:pos], ""
		} else {
			return strings.Join(tokens[:len(tokens)-1], " ") + " ", cs[:pos], ""
		}
	} else {
		cs := []string{"get-oper", "get-conf", "delete", "set", "validate", "commit", "rpc"}
		if len(tokens) > 0 {
			n := 0
			for _, x := range cs {
				if strings.HasPrefix(x, tokens[0]) {
					cs[n] = x
					n++
				}
			}
			cs = cs[:n]
		}
		return "", cs, ""
	}
}

var testMode = false

func main() {
	var port *int
	var port1 int
	var addr1, logLevel1 string
	var addr, logLevel *string
	if flag.CommandLine.Lookup("port") != nil {
		// port = flag.Getter(flag.CommandLine.Lookup("port").Value).Get()

		port1 = flag.Lookup("port").Value.(flag.Getter).Get().(int)
		port = &port1

		addr1 = flag.Lookup("address").Value.(flag.Getter).Get().(string)
		addr = &addr1
		logLevel1 = flag.Lookup("debug").Value.(flag.Getter).Get().(string)
		logLevel = &logLevel1
	} else {
		// Parse args
		port = flag.Int("port", 10555, "Port number to connect to")
		addr = flag.String("address", "localhost", "Address or host to connect to")
		logLevel = flag.String("debug", log.InfoLevel.String(), "debug level")
		flag.Parse()
	}

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

	if !testMode {
		defer s.Close()
	}

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

	// l, err := readline.NewEx(&readline.Config{
	// 	Prompt:            "netconf> ",
	// 	HistoryFile:       "/tmp/readline.tmp",
	// 	AutoComplete:      completer,
	// 	InterruptPrompt:   "^C",
	// 	EOFPrompt:         "exit",
	// 	HistorySearchFold: true,
	// })
	// if err != nil {
	// 	println("Error!")
	// 	panic(err)
	// }
	// defer l.Close()
	// log.SetOutput(l.Stderr())
	var requestLine string

	var liner2 *liner.State = liner.NewLiner()
	defer liner2.Close()
	liner2.SetWordCompleter(wordCompleter)
	liner2.SetTabCompletionStyle(liner.TabPrints)
	if f, err := os.Open(historyFile); err == nil {
		liner2.ReadHistory(f)
		f.Close()
	}
	defer func() {
		if f, err := os.Create(historyFile); err != nil {
			log.Print("Error writing history file: ", err)
		} else {
			liner2.WriteHistory(f)
			f.Close()
		}
	}()

	if testMode {
		return
	}
	for {
		// Maps string to void
		// Becomes a nested map of strings
		// requestMap := make(map[string]interface{})
		//println("In loop")
		// line, err := l.Readline()
		line, err := liner2.Prompt("netconf> ")
		// fmt.Printf("Liner: %v : %v", line, err)
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
		if len(line) > 0 {
			// Should really be when a command has been validated and executed.
			liner2.AppendHistory(line)
		}
		switch {
		case strings.HasPrefix(line, "set"), strings.HasPrefix(line, "delete"):
			// TODO this is a big current limitation - can only set/delete a
			// single item per-commit
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
		case strings.HasPrefix(line, "get-conf"):
			fallthrough
		case strings.HasPrefix(line, "get-oper"), strings.HasPrefix(line, "rpc"):
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
			var op int
			switch slice[0] {
			case "get-oper":
				op = getOper
			case "rpc":
				// TODO - rpc arg completion not working?
				op = rpcOp
			case "get-conf":
				op = getConf
			}
			netconfData, _ := sendNetconfRequest(s, requestLine, op)
			fmt.Printf("Request data: %v\n", netconfData)
		case strings.HasPrefix(line, "validate"):
			sendNetconfRequest(s, requestLine, validate)
		case strings.HasPrefix(line, "commit"):
			sendNetconfRequest(s, requestLine, commit)
		default:
		}
		log.Debug("you said:", strconv.Quote(line))
	}
}
