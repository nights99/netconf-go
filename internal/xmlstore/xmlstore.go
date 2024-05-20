package xmlstore

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type XMLStore struct {
	Root xmlElement
}

type xmlElementInterface interface {
	insert(path []string)
}

type idRefElement struct {
	xmlElement
	foo string
}

// Custom xml marshalling function for the above type
func (el idRefElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// <type xmlns:idx="urn:ietf:params:xml:ns:yang:iana-if-type">idx:ethernetCsmacd</type>
	// Add the foo attribute to the start element
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "", Local: "xmlns:idx"}, Value: "urn:ietf:params:xml:ns:yang:iana-if-type"})
	start.Name.Local = "type"
	// Encode the start element
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	// Encode the value
	if err := e.EncodeToken(xml.CharData("idx:" + el.foo)); err != nil {
		return err
	}
	// Encode the end element
	return e.EncodeToken(start.End())
}

type xmlElement struct {
	XMLName  xml.Name
	Value    string `xml:",chardata"`
	Children []xmlElementInterface
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
		fmt.Printf("Adding %v to x%vx\n", ss[0], el.XMLName.Local)
		if el.XMLName.Local == "" {
			el.XMLName.Local = ss[0]
			el.Children = append(el.Children, &xmlElement{xml.Name{Space: "", Local: ss[1]}, "", []xmlElementInterface{}})
			if len(ss) == 2 {
				return
			}
			el.Children[len(el.Children)-1].insert(ss[2:])
		} else {
			el.Children = append(el.Children, &xmlElement{xml.Name{Space: "", Local: ss[0]}, "", []xmlElementInterface{}})
			if len(ss) == 1 {
				return
			}
			el.Children[len(el.Children)-1].insert(ss[1:])
		}
	}

	// for _, s := range strings.Split(path, " ") {
	// 	slices.IndexFunc(el.Children, func(el1 xmlElement) bool { return el1.XMLName.Local == s })
	// }

}

// func main() {
// 	foo := xmlElement{xml.Name{Space: "ns1", Local: "foo"}, "val1",
// 		[]xmlElementInterface{
// 			xmlElementInterface(&xmlElement{xml.Name{Space: "ns2", Local: "bar"}, "val2", []xmlElementInterface{}}),
// 			xmlElementInterface(&xmlElement{xml.Name{Space: "", Local: "bar2"}, "", []xmlElementInterface{}}),
// 		},
// 	}
// 	foo.insert(strings.Split("foo bar2 next_level", " "))
// 	foo.Children = append(foo.Children, &idRefElement{xmlElement{xml.Name{Space: "", Local: "next_level"}, "", []xmlElementInterface{}}, "ethernetCsmacd"})
// 	myxml, err := xml.MarshalIndent(foo, "", "  ")
// 	fmt.Printf("%v %v\n", string(myxml), err)
// }

func (store XMLStore) Insert(line string) {
	fmt.Printf("Inserting %v\n", line)
	// Split on space
	ss := strings.Split(line, " ")
	// Drop first element from slice
	ss = ss[1:]
	// Insert into store
	store.Root.insert(ss)
	fmt.Printf("%v\n", store)
	myxml, err := xml.MarshalIndent(store.Root, "", "  ")
	fmt.Printf("%v %v\n", string(myxml), err)
}
