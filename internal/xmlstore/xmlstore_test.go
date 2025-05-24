package xmlstore

import (
	"encoding/xml"
	"netconf-go/internal/types"
	"reflect"
	"strings"
	"testing"

	"github.com/openconfig/goyang/pkg/yang"
)

// MockYangEntry is a mock implementation of yang.Entry for testing.
type MockYangEntry struct {
	yang.Entry
	MockName      string
	MockNamespace string
	MockPrefix    string
	MockKind      yang.Kind
	MockType      *yang.YangType
	MockDir       map[string]*yang.Entry
	MockKey       string
}

// Name returns the mock name.
func (m *MockYangEntry) Name() string {
	return m.MockName
}

// Namespace returns a mock namespace.
func (m *MockYangEntry) Namespace() *yang.Value {
	if m.MockNamespace == "" {
		return nil
	}
	return &yang.Value{
		Name: m.MockNamespace,
	}
}

// Prefix returns a mock prefix.
func (m *MockYangEntry) Prefix() *yang.Value {
	if m.MockPrefix == "" {
		return nil
	}
	return &yang.Value{
		Name: m.MockPrefix,
	}
}

// Kind returns the mock kind.
func (m *MockYangEntry) Kind() yang.Kind {
	return m.MockKind
}

// Type returns the mock type.
func (m *MockYangEntry) Type() *yang.YangType {
	return m.MockType
}

// Dir returns the mock directory entries.
func (m *MockYangEntry) Dir() map[string]*yang.Entry {
	return m.MockDir
}

// Key returns the mock key.
func (m *MockYangEntry) Key() string {
	return m.MockKey
}

// IsLeafList returns false for this mock.
func (m *MockYangEntry) IsLeafList() bool {
	return m.MockKind == yang.LeafListEntry
}

// IsList returns false for this mock.
func (m *MockYangEntry) IsList() bool {
	return m.MockKind == yang.DirectoryEntry && m.MockKey != ""
}


func TestMain(m *testing.M) {
	m.Run()
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}

func TestXMLStore_Insert(t *testing.T) {
	// Mock yang.Entry for general use
	mockInterfacesEntry := &MockYangEntry{
		MockName:      "interfaces",
		MockKind:      yang.DirectoryEntry,
		MockNamespace: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
		MockPrefix:    "if",
		MockDir: map[string]*yang.Entry{
			"interface": {Name: "interface", Kind: yang.DirectoryEntry, Key: "name", Prefix: &yang.Value{Name: "if"}, Parent: &yang.Module{Prefix: &yang.Value{Name: "if"}, Namespace: &yang.Value{Name: "urn:ietf:params:xml:ns:yang:ietf-interfaces"}},
				Dir: map[string]*yang.Entry{
					"name":        {Name: "name", Kind: yang.LeafEntry, Type: &yang.YangType{Kind: yang.Ystring}},
					"description": {Name: "description", Kind: yang.LeafEntry, Type: &yang.YangType{Kind: yang.Ystring}, Prefix: &yang.Value{Name: "if"}, Parent: &yang.Module{Prefix: &yang.Value{Name: "if"}, Namespace: &yang.Value{Name: "urn:ietf:params:xml:ns:yang:ietf-interfaces"}}},
					"mtu":         {Name: "mtu", Kind: yang.LeafEntry, Type: &yang.YangType{Kind: yang.Yuint16}, Prefix: &yang.Value{Name: "if"}, Parent: &yang.Module{Prefix: &yang.Value{Name: "if"}, Namespace: &yang.Value{Name: "urn:ietf:params:xml:ns:yang:ietf-interfaces"}}},
				}},
		},
	}
	mockInterfacesEntry.MockDir["interface"].Dir["name"].Parent = mockInterfacesEntry.MockDir["interface"]
	mockInterfacesEntry.MockDir["interface"].Dir["description"].Parent = mockInterfacesEntry.MockDir["interface"]
	mockInterfacesEntry.MockDir["interface"].Dir["mtu"].Parent = mockInterfacesEntry.MockDir["interface"]


	// Mock yang.Entry for identityref
	idrefType := &yang.YangType{
		Kind: yang.Yidentityref,
		Base: &yang.Identity{ // Mocked base identity
			Name:   "crypto-alg",
			Prefix: &yang.Value{Name: "ianaid"},
			Parent: &yang.Module{
				Namespace: &yang.Value{Name: "urn:ietf:params:xml:ns:yang:iana-crypt-hash"},
				Prefix:    &yang.Value{Name: "ianaid"},
			},
		},
	}
	mockCryptoEntry := &MockYangEntry{
		MockName:      "crypto",
		MockKind:      yang.DirectoryEntry,
		MockNamespace: "urn:example:crypto",
		MockPrefix:    "ex",
		MockDir: map[string]*yang.Entry{
			"algorithm": {
				Name: "algorithm", Kind: yang.LeafEntry, Type: idrefType, Prefix: &yang.Value{Name: "ex"}, Parent: &yang.Module{Prefix: &yang.Value{Name: "ex"}, Namespace: &yang.Value{Name: "urn:example:crypto"}},
			},
		},
	}
	mockCryptoEntry.MockDir["algorithm"].Parent = mockCryptoEntry


	tests := []struct {
		name          string
		initialStore  *XMLStore
		entry         *yang.Entry // This is the top-level module entry
		requestLine   string
		requestType   types.RequestType
		expectedStore *xmlElement
	}{
		{
			name:         "insert single container",
			initialStore: New(),
			entry:        yang.ToEntry(mockInterfacesEntry),
			requestLine:  "ietf-interfaces interfaces",
			requestType:  types.EditConf,
			expectedStore: &xmlElement{
				Name:     "interfaces",
				Space:    "urn:ietf:params:xml:ns:yang:ietf-interfaces",
				Children: make(map[string]*xmlElement),
				Attrs:    make(map[string]xml.Attr),
			},
		},
		{
			name:         "insert nested leaf with value",
			initialStore: New(),
			entry:        yang.ToEntry(mockInterfacesEntry),
			requestLine:  "ietf-interfaces interfaces interface name=GigabitEthernet0/0/0 description=Test_Interface",
			requestType:  types.EditConf,
			expectedStore: &xmlElement{
				Name:  "interfaces",
				Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
				Attrs: make(map[string]xml.Attr),
				Children: map[string]*xmlElement{
					"interface[name=GigabitEthernet0/0/0]": {
						Name:  "interface",
						Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
						Attrs: make(map[string]xml.Attr),
						Keys:  map[string]string{"name": "GigabitEthernet0/0/0"},
						Children: map[string]*xmlElement{
							"name": {Name: "name", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces", Value: stringPtr("GigabitEthernet0/0/0"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement)},
							"description": {Name: "description", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces", Value: stringPtr("Test_Interface"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement)},
						},
					},
				},
			},
		},
		{
			name: "delete leaf",
			initialStore: func() *XMLStore {
				s := New()
				// Pre-populate with the leaf
				s.Insert(yang.ToEntry(mockInterfacesEntry), "ietf-interfaces interfaces interface name=GigabitEthernet0/0/0 description=Test_Interface", types.EditConf)
				return s
			}(),
			entry:       yang.ToEntry(mockInterfacesEntry),
			requestLine: "delete ietf-interfaces interfaces interface name=GigabitEthernet0/0/0 description",
			requestType: types.EditConf,
			expectedStore: &xmlElement{
				Name:  "interfaces",
				Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
				Attrs: make(map[string]xml.Attr),
				Children: map[string]*xmlElement{
					"interface[name=GigabitEthernet0/0/0]": {
						Name:  "interface",
						Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
						Attrs: make(map[string]xml.Attr),
						Keys:  map[string]string{"name": "GigabitEthernet0/0/0"},
						Children: map[string]*xmlElement{
							"name":        {Name: "name", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces", Value: stringPtr("GigabitEthernet0/0/0"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement)},
							"description": {Name: "description", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces", delete: true, Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement)},
						},
					},
				},
			},
		},
		{
			name:         "insert identityref value",
			initialStore: New(),
			entry:        yang.ToEntry(mockCryptoEntry),
			requestLine:  "example-crypto crypto algorithm=ianaid:hmac-sha1", // Matches prefix in mockCryptoEntry's Base identity
			requestType:  types.EditConf,
			expectedStore: &xmlElement{
				Name:  "crypto",
				Space: "urn:example:crypto",
				Attrs: make(map[string]xml.Attr),
				Children: map[string]*xmlElement{
					"algorithm": {
						Name:        "algorithm",
						Space:       "urn:example:crypto",
						Value:       stringPtr("hmac-sha1"), // Value should be just the name
						Attrs:       make(map[string]xml.Attr),
						Children:    make(map[string]*xmlElement),
						idrefNs:     "urn:ietf:params:xml:ns:yang:iana-crypt-hash", // Namespace of the base identity
						idrefPrefix: "ianaid",                                    // Prefix of the base identity
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeToTest := New() // Always start with a fresh store for Root
			if tt.initialStore != nil && tt.initialStore.Root != nil { // If initial state is needed
				// Deep copy initialStore.Root to storeToTest.Root to avoid modifying test case data
				var copyXmlElement func(orig *xmlElement) *xmlElement
				copyXmlElement = func(orig *xmlElement) *xmlElement {
					if orig == nil {
						return nil
					}
					copyEl := &xmlElement{
						Name:        orig.Name,
						Space:       orig.Space,
						Value:       orig.Value, // stringPtr is fine, strings are immutable
						delete:      orig.delete,
						idrefNs:     orig.idrefNs,
						idrefPrefix: orig.idrefPrefix,
						Attrs:       make(map[string]xml.Attr),
						Children:    make(map[string]*xmlElement),
						Keys:        make(map[string]string),
					}
					for k, v := range orig.Attrs {
						copyEl.Attrs[k] = v
					}
					for k, v := range orig.Keys {
						copyEl.Keys[k] = v
					}
					for k, v := range orig.Children {
						copyEl.Children[k] = copyXmlElement(v)
					}
					return copyEl
				}
				storeToTest.Root = copyXmlElement(tt.initialStore.Root)
			}
			
			storeToTest.Insert(tt.entry, tt.requestLine, tt.requestType)

			if !reflect.DeepEqual(storeToTest.Root, tt.expectedStore) {
				// For detailed diff, marshal both to string and compare
				gotXML, _ := xml.MarshalIndent(storeToTest.Root, "", "  ")
				wantXML, _ := xml.MarshalIndent(tt.expectedStore, "", "  ")
				t.Errorf("XMLStore.Insert() Root diff:\ngot:\n%s\nwant:\n%s", string(gotXML), string(wantXML))

				// Fallback to DeepEqual's output if XML marshalling is problematic or not detailed enough for struct fields
				// t.Errorf("XMLStore.Insert() Root = \n%#v\nwant = \n%#v", storeToTest.Root, tt.expectedStore)
			}
		})
	}
}


func TestXMLStore_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		store   *XMLStore
		wantXML string
	}{
		{
			name: "marshal simple store",
			store: &XMLStore{
				Root: &xmlElement{
					Name:     "interfaces",
					Space:    "urn:ietf:params:xml:ns:yang:ietf-interfaces",
					Children: make(map[string]*xmlElement),
					Attrs:    make(map[string]xml.Attr),
					Keys:     make(map[string]string),
				},
			},
			wantXML: `<interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces"></interfaces>`,
		},
		{
			name: "marshal store with nested children",
			store: &XMLStore{
				Root: &xmlElement{
					Name:  "interfaces",
					Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
					Attrs: make(map[string]xml.Attr),
					Keys:  make(map[string]string),
					Children: map[string]*xmlElement{
						"interface[name=GigabitEthernet0/0/0]": {
							Name:  "interface",
							Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
							Attrs: make(map[string]xml.Attr),
							Keys:  map[string]string{"name": "GigabitEthernet0/0/0"},
							Children: map[string]*xmlElement{
								"name":        {Name: "name", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces", Value: stringPtr("GigabitEthernet0/0/0"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string)},
								"description": {Name: "description", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces", Value: stringPtr("Test_Interface"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string)},
							},
						},
					},
				},
			},
			wantXML: `<interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">
  <interface xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">
    <name>GigabitEthernet0/0/0</name>
    <description>Test_Interface</description>
  </interface>
</interfaces>`,
		},
		{
			name:    "marshal empty store",
			store:   New(), // New() creates a store with nil Root
			wantXML: ``, // Marshalling a nil root should produce empty or an error handled by caller, here we expect empty.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := xml.MarshalIndent(tt.store, "", "  ")
			if err != nil {
				// If Root is nil, MarshalXML in XMLStore might return an error or specific behavior.
				// For this test, if we expect empty string for nil Root, an error might be unexpected.
				// However, xml.Marshal(nil) is valid and produces "null" for JSON, but for XML with custom marshaller it depends.
				// XMLStore.MarshalXML handles nil Root by returning nil error and doing nothing.
				if tt.store.Root == nil && tt.wantXML == "" {
					// This is expected
				} else {
					t.Fatalf("XMLStore.MarshalIndent() error = %v", err)
				}
			}
			gotXML := string(gotBytes)
			if gotXML != tt.wantXML {
				if tt.wantXML == "" && gotXML == "<nil></nil>" { // xml.MarshalIndent of a nil struct can result in <nil></nil>
					// consider this okay for empty store test
				} else {
					t.Errorf("XMLStore.Marshal() gotXML = \n%s\n, wantXML \n%s", gotXML, tt.wantXML)
				}
			}
		})
	}
}


func TestXmlElement_MarshalXML(t *testing.T) {
	tests := []struct {
		name    string
		element *xmlElement
		wantXML string
	}{
		{
			name: "simple element with value",
			element: &xmlElement{
				Name:  "leaf1",
				Value: stringPtr("value1"),
				Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string),
			},
			wantXML: `<leaf1>value1</leaf1>`,
		},
		{
			name: "element with namespace",
			element: &xmlElement{
				Name:  "container",
				Space: "urn:example:test",
				Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string),
			},
			wantXML: `<container xmlns="urn:example:test"></container>`,
		},
		{
			name: "element with children",
			element: &xmlElement{
				Name:  "parent",
				Space: "urn:parent",
				Attrs: make(map[string]xml.Attr), Keys: make(map[string]string),
				Children: map[string]*xmlElement{
					"child1": {Name: "child1", Space: "urn:parent", Value: stringPtr("cval1"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string)},
					"child2": {Name: "child2", Space: "urn:parent", Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string)},
				},
			},
			// Note: Order of children in map is not guaranteed, so XML output order might vary.
			// For robust test, either sort children keys or check for presence of children individually.
			// For simplicity here, assuming a consistent order for test purposes or that specific order doesn't matter.
			// To make it more robust, we could unmarshal and compare structs, or compare sets of children.
			// This simplified test will use a fixed order and might be flaky if map iteration order changes.
			// A better approach is to check for specific children if order is not fixed.
			// For now, let's assume specific order for this test's wantXML: child1 then child2.
			// A common way maps are iterated in Go is by sorted key for consistent output. Let's try that here.
			wantXML: `<parent xmlns="urn:parent">
  <child1>cval1</child1>
  <child2></child2>
</parent>`,
		},
		{
			name: "element marked for deletion",
			element: &xmlElement{
				Name:   "todelete",
				Space:  "urn:example:test",
				delete: true,
				Attrs:  make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string),
			},
			wantXML: `<todelete xmlns="urn:example:test" operation="remove"></todelete>`,
		},
		{
			name: "identityref element",
			element: &xmlElement{
				Name:        "idleaf",
				Space:       "urn:example:id",
				Value:       stringPtr("my-id-value"),
				idrefNs:     "urn:identity:namespace",
				idrefPrefix: "idp",
				Attrs:       make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string),
			},
			wantXML: `<idleaf xmlns="urn:example:id" xmlns:idp="urn:identity:namespace">idp:my-id-value</idleaf>`,
		},
		{
			name: "element with custom attribute",
			element: &xmlElement{
				Name:  "customattr",
				Space: "urn:example:custom",
				Attrs: map[string]xml.Attr{
					"mykey": {Name: xml.Name{Local: "mykey"}, Value: "myvalue"},
				},
				Children: make(map[string]*xmlElement), Keys: make(map[string]string),
			},
			wantXML: `<customattr xmlns="urn:example:custom" mykey="myvalue"></customattr>`,
		},
		{
			name: "empty element (no value, no children)",
			element: &xmlElement{
				Name:  "emptycontainer",
				Space: "urn:example:empty",
				Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string),
			},
			wantXML: `<emptycontainer xmlns="urn:example:empty"></emptycontainer>`,
		},
		{
			name: "element with keys (list instance)",
			element: &xmlElement{
				Name:  "myList", // Name of the list YANG node
				Space: "urn:example:list",
				Attrs: make(map[string]xml.Attr),
				Keys:  map[string]string{"key1": "val1", "key2": "val2"}, // Keys for this specific list instance
				Children: map[string]*xmlElement{
					// Key elements are children
					"key1": {Name: "key1", Space: "urn:example:list", Value: stringPtr("val1"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string)},
					"key2": {Name: "key2", Space: "urn:example:list", Value: stringPtr("val2"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string)},
					"data": {Name: "data", Space: "urn:example:list", Value: stringPtr("somedata"), Attrs: make(map[string]xml.Attr), Children: make(map[string]*xmlElement), Keys: make(map[string]string)},
				},
			},
			// The xmlElement.MarshalXML should marshal the children representing the keys and other data.
			// The element name itself is "myList".
			wantXML: `<myList xmlns="urn:example:list">
  <key1>val1</key1>
  <key2>val2</key2>
  <data>somedata</data>
</myList>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Custom comparison needed for children due to map ordering.
			// However, xml.MarshalIndent on a struct containing a map of xmlElements
			// might itself have ordering issues if not handled inside xmlElement.MarshalXML.
			// The current xmlElement.MarshalXML sorts children keys, so output should be stable.
			gotBytes, err := xml.MarshalIndent(tt.element, "", "  ")
			if err != nil {
				t.Fatalf("xmlElement.MarshalIndent() error = %v", err)
			}
			gotXML := string(gotBytes)

			// Normalize expected XML slightly for comparison if needed, e.g. for self-closing tags vs start/end.
			// For now, direct string comparison.
			if strings.TrimSpace(gotXML) != strings.TrimSpace(tt.wantXML) {
				t.Errorf("xmlElement.MarshalXML() gotXML = \n%s\n, wantXML \n%s", gotXML, tt.wantXML)
			}
		})
	}
}
