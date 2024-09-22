package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

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

func (el xmlElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = el.XMLName.Local
	start.Name.Space = el.XMLName.Space
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	e.EncodeElement(el.Children, start)

	return e.EncodeToken(start.End())
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
		el.Children = append(el.Children, &xmlElement{xml.Name{Local: ss[1], Space: ""}, "", []xmlElementInterface{}})
	}

	// for _, s := range strings.Split(path, " ") {
	// 	slices.IndexFunc(el.Children, func(el1 xmlElement) bool { return el1.XMLName.Local == s })
	// }

}

func main() {
	foo := xmlElement{xml.Name{Space: "ns1", Local: "foo"}, "val1",
		[]xmlElementInterface{
			xmlElementInterface(&xmlElement{xml.Name{Space: "ns2", Local: "bar"}, "val2", []xmlElementInterface{}}),
			xmlElementInterface(&xmlElement{xml.Name{Space: "", Local: "bar2"}, "", []xmlElementInterface{}}),
		},
	}
	foo.insert(strings.Split("foo bar2 next_level", " "))
	foo.Children = append(foo.Children, &idRefElement{xmlElement{xml.Name{Space: "ns2", Local: "next_level2"}, "", []xmlElementInterface{}}, "ethernetCsmacd"})
	myxml, err := xml.MarshalIndent(foo, "", "  ")
	fmt.Printf("%v %v\n", string(myxml), err)
}
