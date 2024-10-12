package xmlstore

import (
	"encoding/xml"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

type XMLStore struct {
	Root xmlElement
}

type xmlElementInterface interface {
	insert(yangMod *yang.Entry, path []string)
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
func (el *xmlElement) insert(yangMod *yang.Entry, path []string) {
	// ss := strings.Split(path, " ")
	ss := path
	if ss[0] == el.XMLName.Local {
		// Found element, recurse
		// println("Found ", ss[0])
		for i := range el.Children {
			el.Children[i].insert(nil, ss[1:])
		}
	} else {
		// Add new element, then insert into that
		// fmt.Printf("Adding %v to x%vx\n", ss[0], el.XMLName.Local)
		if el.XMLName.Local == "" {
			el.XMLName.Local = ss[0]
			el.XMLName.Space = yangMod.Namespace().Name
			el.Children = append(el.Children, &xmlElement{xml.Name{Space: "", Local: ss[1]}, "", []xmlElementInterface{}})
			if len(ss) == 2 {
				return
			}
			el.Children[len(el.Children)-1].insert(nil, ss[2:])
		} else {
			el.Children = append(el.Children, &xmlElement{xml.Name{
				Local: ss[0],
				// Space: el.XMLName.Space
			},
				"", []xmlElementInterface{}})
			if len(ss) == 1 {
				return
			}
			el.Children[len(el.Children)-1].insert(nil, ss[1:])
		}
	}
}

func (store *XMLStore) Insert(yangMod *yang.Entry, line string) {
	// fmt.Printf("Inserting %v\n", line)
	// Split on space
	ss := strings.Split(line, " ")
	// Drop first element from slice
	ss = ss[1:]
	// Insert into store
	store.Root.XMLName.Local = ss[1]
	store.Root.XMLName.Space = yangMod.Namespace().Name
	if len(ss) > 2 {
		store.Root.insert(yangMod, ss[2:])
	}
	// fmt.Printf("%v\n", store)
	// myxml, err := xml.MarshalIndent(store.Root, "", "  ")
	// fmt.Printf("%v %v\n", string(myxml), err)
}

func (store *XMLStore) Marshal() string {
	myxml, _ := xml.MarshalIndent(store.Root, "", "  ")
	// fmt.Printf("%v %v\n", string(myxml), err)
	return string(myxml)
}
