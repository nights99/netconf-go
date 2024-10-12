//go:build !wasm
// +build !wasm

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	// "netconf-go/internal/lib"

	// "netconf-go/internal/lib"

	"github.com/chzyer/readline"
	// netconf "github.com/nemith/go-netconf/v2"
	// ncssh "github.com/nemith/go-netconf/v2/transport/ssh"
	netconf "github.com/nemith/netconf"
	ncssh "github.com/nemith/netconf/transport/ssh"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/peterh/liner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

// Array of available Yang modules
// TODO dubious sharing of global variables with the lib file...
var modNames []string

var historyFile = filepath.Join(os.TempDir(), ".liner_example_history")

func wordCompleter(line string, pos2 int) (head string, completions []string, tail string) {
	// fmt.Printf("Called with %v\n", line)

	// TODO check if pos is end of line

	tokens := strings.Fields(line)
	log.Debugf("tokens: %d, %v", len(tokens), tokens)

	if len(tokens) >= 2 || strings.HasSuffix(line, " ") {
		yangCompletions, returnType := ListYang(line)
		// fmt.Printf("Completions %v\n", yangCompletions)

		cs := make([]string, len(yangCompletions))
		var pos int = 0
		var found_augment bool = false
		for _, e := range yangCompletions {
			// fmt.Printf("Comparing '%s' and '%s'\n", e[strings.LastIndex(e, " ")+1:], tokens[len(tokens)-1])
			if strings.Contains(e, "@") {
				// Needs quite different handling - need to return a start containing these tokens.
				if len(yangCompletions) == 1 {
					var prefix = strings.Join(append([]string{tokens[0] + " "}, yangCompletions[0:len(yangCompletions)-1]...), " ")
					return prefix, []string{yangCompletions[len(yangCompletions)-1]}, ""
				}
			}
			if strings.Contains(e, " ") {
				found_augment = true
				tokens2 := strings.Split(e, " ")
				if strings.HasPrefix(tokens2[len(tokens2)-1], tokens[len(tokens)-1]) ||
					strings.HasSuffix(tokens2[len(tokens2)-2], tokens[len(tokens)-1]) || // Handles case where we have an augment and then a new token after it. Suffix isn't really right - we want to strip the augment, then compare the prefix
					strings.HasSuffix(line, " ") ||
					strings.HasPrefix(tokens[len(tokens)-1], "/") {
					cs[pos] = tokens2[len(tokens2)-1]
					pos++
				}
			} else if strings.HasPrefix(e, tokens[len(tokens)-1]) || strings.HasSuffix(line, " ") {
				cs[pos] = e
				pos++
			}
		}
		// cs = []string{longestcommon.Prefix(cs[:pos])}
		// fmt.Printf("Found %v\n", cs)
		// Add a space on the end if we've found a completion
		if len(cs[:pos]) == 1 && returnType != replaceLast {
			cs[0] += " "
		}
		if returnType == replaceLast {
			tokens = tokens[:len(tokens)-1]
		}
		if strings.HasSuffix(line, " ") || returnType == replaceLast {
			return strings.Join(tokens, " ") + " ", cs[:pos], ""
		} else {
			if found_augment {
				tokens2 := strings.Split(yangCompletions[0], " ")
				var prefix = strings.Join(append([]string{tokens[0]}, tokens2[0:len(tokens2)-1]...), " ")
				return prefix + " ", cs[:pos], ""
			} else {
				return strings.Join(tokens[:len(tokens)-1], " ") + " ", cs[:pos], ""
			}
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

func main() {
	var port int
	var addr string
	var logLevel *string
	var telnet *bool
	var testMode = false
	var err error

	// Read config file using Viper
	viper.SetConfigName("hosts")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			panic(err)
		}
	}

	// Config file found and successfully parsed
	// Parse args
	pflag.Int("port", 22, "Port number to connect to")
	pflag.String("address", "localhost", "Address or host to connect to")
	host := pflag.String("host", "", "Hostname referring to hosts.yaml entry")
	telnet = pflag.Bool("t", false, "Use telnet to connect")
	logLevel = pflag.String("debug", log.InfoLevel.String(), "debug level")

	pflag.String("user", "", "Username")
	pflag.String("password", "", "Password")

	pflag.Parse()
	var hostConfig *viper.Viper
	if *host != "" {
		hostConfig = viper.Sub(*host)
		if hostConfig == nil { // Sub returns nil if the key cannot be found
			panic("host configuration not found")
		}
		hostConfig.BindPFlags(pflag.CommandLine)
	}
	viper.BindPFlags(pflag.CommandLine)

	log.Debugf("pflags: %v\n", viper.AllSettings())
	if hostConfig != nil {
		log.Debugf("host pflags: %v\n", hostConfig.AllSettings())
		viper.MergeConfigMap(hostConfig.AllSettings())
	}
	log.Debugf("merge pflags: %v\n", viper.AllSettings())
	addr = viper.GetString("address")
	port = viper.GetInt("port")
	user := viper.GetString("user")
	password := viper.GetString("password")

	l2, _ := log.ParseLevel(*logLevel)
	log.SetLevel(l2)

	// Connect to the node
	var s *netconf.Session
	if *telnet {
		transport, err := DialTelnet(addr+":"+strconv.Itoa(port), "lab", "lab", nil)
		if err != nil {
			panic(err)
		}
		s, err = netconf.Open(transport)
		if err != nil {
			panic(err)
		}
	} else {
		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		// s, err = netconf.DialSSH(*addr+":"+strconv.Itoa(*port), sshConfig)

		transport, err := ncssh.Dial(context.Background(), "tcp", addr+":"+strconv.Itoa(port), sshConfig)
		if err != nil {
			panic(err)
		}
		defer transport.Close()

		s, err = netconf.Open(transport)
		if err != nil {
			panic(err)
		}
	}

	if err != nil {
		panic(err)
	}
	globalSession = s

	if !testMode {
		defer s.Close(context.Background())
	}

	// fmt.Printf("Server Capabilities: '%+v'\n", s.ServerCapabilities[0])
	// fmt.Printf("Session Id: %d\n\n", s.SessionID)

	ms = yang.NewModules()

	modNames = getSchemaList(s)
	//fmt.Printf("modNames: %v\n", modNames)
	sort.Strings(modNames)
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
	getYangModule(s, "iana-if-type")
	for {
		// Maps string to void
		// Becomes a nested map of strings
		// requestMap := make(map[string]interface{})
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
			netconfData, _ := sendNetconfRequest(s, requestLine, editConf)
			fmt.Printf("Request data: %v\n", netconfData)
		case strings.HasPrefix(line, "get-conf"):
			// TODO make common with get-oper/rpc below using fallthrough
			requestLine = line
			slice := strings.Split(requestLine, " ")
			log.Debug("Set line:", slice[1:])

			/*
			 * If we don't know the module, read it from the router now.
			 */
			if len(slice) >= 2 && mods[slice[1]] == nil {
				mods[slice[1]] = getYangModule(s, slice[1])
				if mods[slice[1]] == nil {
					continue
				}
			}
			netconfData, _ := sendNetconfRequest(s, requestLine, getConf)
			fmt.Printf("Request data: %v\n", netconfData)
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
			var op requestType
			switch slice[0] {
			case "get-oper":
				op = getOper
			case "rpc":
				op = rpcOp
			case "get-conf":
				op = getConf
			}
			netconfData, _ := sendNetconfRequest(s, requestLine, op)
			fmt.Printf("Request data: %v\n", netconfData)
		case strings.HasPrefix(line, "validate"):
			requestLine = line
			sendNetconfRequest(s, requestLine, validate)
		case strings.HasPrefix(line, "commit"):
			requestLine = line
			sendNetconfRequest(s, requestLine, commit)
		default:
		}
		log.Debug("you said:", strconv.Quote(line))
	}
}
