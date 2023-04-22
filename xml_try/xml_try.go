package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type xmlElement struct {
	XMLName  xml.Name
	Value    string `xml:",chardata"`
	Children []xmlElement
}

func (el *xmlElement) insert(path []string) {
	// ss := strings.Split(path, " ")
	ss := path
	if ss[0] == el.XMLName.Local {
		// Found element, recurse
		println("Found ", ss[0])
		for i := range el.Children {
			el.Children[i].insert(ss[1:])
		}
	} else {
		// Add new element, then insert into that
		el.Children = append(el.Children, xmlElement{xml.Name{Space: "", Local: ss[1]}, "", []xmlElement{}})
	}

	// for _, s := range strings.Split(path, " ") {
	// 	slices.IndexFunc(el.Children, func(el1 xmlElement) bool { return el1.XMLName.Local == s })
	// }

}

func main() {
	foo := xmlElement{xml.Name{Space: "ns1", Local: "foo"}, "val1",
		[]xmlElement{
			{xml.Name{Space: "ns2", Local: "bar"}, "val2", []xmlElement{}},
			{xml.Name{Space: "", Local: "bar2"}, "", []xmlElement{}},
		},
	}
	foo.insert(strings.Split("foo bar2 next_level", " "))
	myxml, err := xml.MarshalIndent(foo, "", "  ")
	fmt.Printf("%v %v\n", string(myxml), err)
}
