package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/openconfig/goyang/pkg/yang"
)

var completer = readline.NewPrefixCompleter(readline.PcItem("get"))

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

	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[31m»\033[0m ",
		HistoryFile:       "/tmp/readline.tmp",
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	log.SetOutput(l.Stderr())
	for {
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
		case strings.HasPrefix(line, "get "):
		default:
			log.Println("you said:", strconv.Quote(line))
		}
	}
}
