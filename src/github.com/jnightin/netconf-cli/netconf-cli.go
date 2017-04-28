package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/openconfig/goyang/pkg/yang"
)

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
}
