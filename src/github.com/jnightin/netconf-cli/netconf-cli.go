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

var completer = readline.NewPrefixCompleter(readline.PcItem("get", readline.PcItemDynamic(listYang("./"))))

func listYang(path string) func(string) []string {
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
				println("Error!")
				println("Error!")
				println("Error!")
				println("Error!")
				println("Error!")
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			println("Error!")
			println("Error!")
			println("Error!")
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "get"):
		default:
			log.Println("you said:", strconv.Quote(line))
		}
	}

}
