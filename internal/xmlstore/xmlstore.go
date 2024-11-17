package xmlstore

import (
	"encoding/xml"
	"netconf-go/internal/types"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

type XMLStore struct {
	Root xmlElement
}

type xmlElementInterface interface {
	insert(yangMod *yang.Entry, path []string, requestType types.RequestType, delete bool)
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
	delete   bool
	idref    *string
}

func (el xmlElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = el.XMLName.Local
	start.Name.Space = el.XMLName.Space
	if el.delete {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "operation", Space: "urn:ietf:params:xml:ns:netconf:base:1.0"}, Value: "remove"})
	}
	if el.idref != nil {
		// "urn:ietf:params:xml:ns:yang:iana-if-type"
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "", Local: "xmlns:idx"}, Value: *el.idref})
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if el.Value != "" {
		if el.idref != nil {
			e.EncodeToken(xml.CharData("idx:" + el.Value))
		} else {
			e.EncodeToken(xml.CharData(el.Value))
		}
	} else {
		e.EncodeElement(el.Children, start)
	}

	return e.EncodeToken(start.End())
}
func (el *xmlElement) insert(yangMod *yang.Entry, path []string, requestType types.RequestType, delete bool) {
	var done bool = false
	// ss := strings.Split(path, " ")
	ss := path
	if ss[0] == el.XMLName.Local {
		// Found element, recurse
		// println("Found ", ss[0])
		for i := range el.Children {
			el.Children[i].insert(nil, ss[1:], requestType, delete)
		}
	} else {
		// Add new element, then insert into that
		// fmt.Printf("Adding %v to x%vx\n", ss[0], el.XMLName.Local)
		nv := strings.Split(ss[0], "=")
		if el.XMLName.Local == "" {
			el.XMLName.Local = nv[0]
			if len(nv) > 1 {
				el.Value = nv[1]
			}
			el.XMLName.Space = yangMod.Namespace().Name
			el.Children = append(el.Children, &xmlElement{xml.Name{Space: "", Local: ss[1]}, "", []xmlElementInterface{}, false, nil})
			if len(ss) == 2 {
				return
			}
			el.Children[len(el.Children)-1].insert(nil, ss[2:], requestType, delete)
		} else {
			child := xmlElement{xml.Name{
				Local: nv[0],
				// Space: el.XMLName.Space
			}, "", []xmlElementInterface{}, false, nil}
			if strings.Contains(nv[0], "@") {
				spaceLocal := strings.Split(nv[0], "@")
				child.XMLName.Space = spaceLocal[0]
				child.XMLName.Local = spaceLocal[1]
			}
			if len(nv) > 1 {
				child.Value = nv[1]
			}
			if len(path) == 1 && requestType == types.EditConf && delete {
				child.delete = true
			}
			if len(path) == 2 && requestType == types.EditConf && !delete && strings.Contains(path[1], "@") {
				// Set idref to be the part before the @
				idref := strings.Split(path[1], "@")[0]
				child.idref = &idref
				child.Value = strings.Split(path[1], "@")[1]
				el.Children = append(el.Children, &child)
				done = true
			} else if len(path) == 1 && requestType == types.EditConf && !delete {
				// If its a set, the last path element is the value.
				el.Value = path[0]
			} else {
				el.Children = append(el.Children, &child)
			}
			if len(ss) == 1 || done {
				return
			}
			if len(nv) > 1 {
				el.insert(nil, ss[1:], requestType, delete)
			} else {
				el.Children[len(el.Children)-1].insert(nil, ss[1:], requestType, delete)

			}
		}
	}
}

func (store *XMLStore) Insert(yangMod *yang.Entry, line string, requestType types.RequestType) {
	// fmt.Printf("Inserting %v\n", line)
	// Split on space
	ss := strings.Split(line, " ")
	isDelete := ss[0] == "delete"
	// Drop first element from slice
	ss = ss[1:]
	// Insert into store
	store.Root.XMLName.Local = ss[1]
	store.Root.XMLName.Space = yangMod.Namespace().Name
	if len(ss) > 2 {
		store.Root.insert(yangMod, ss[2:], requestType, isDelete)
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
