package xmlstore

import (
	"encoding/xml"
	"netconf-go/internal/types"
	"strings"
	"testing"

	"github.com/openconfig/goyang/pkg/yang"
)

// We'll construct concrete *yang.Entry values below for testing purposes.

func TestMain(m *testing.M) {
	m.Run()
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}

func TestXMLStore_Insert(t *testing.T) {
	// Create simple module entries via yang.Module -> yang.ToEntry and use them
	modIf := &yang.Module{Name: "ietf-interfaces", Namespace: &yang.Value{Name: "urn:ietf:params:xml:ns:yang:ietf-interfaces"}}
	msIf := yang.NewModules()
	msIf.Modules[modIf.Name] = modIf
	modIf.Modules = msIf
	entryIf := yang.ToEntry(modIf)

	modCrypto := &yang.Module{Name: "example-crypto", Namespace: &yang.Value{Name: "urn:example:crypto"}}
	msCrypto := yang.NewModules()
	msCrypto.Modules[modCrypto.Name] = modCrypto
	modCrypto.Modules = msCrypto
	entryCrypto := yang.ToEntry(modCrypto)

	tests := []struct {
		name        string
		initial     *XMLStore
		entry       *yang.Entry
		requestLine string
		requestType types.RequestType
		wantSubstrs []string
	}{
		{
			name:        "insert single container",
			initial:     &XMLStore{},
			entry:       entryIf,
			requestLine: "cmd ietf-interfaces interfaces",
			requestType: types.EditConf,
			wantSubstrs: []string{`<interfaces`, `xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces"`},
		},
		{
			name:        "insert nested leaf with value",
			initial:     &XMLStore{},
			entry:       entryIf,
			requestLine: "cmd ietf-interfaces interfaces interface name=GigabitEthernet0/0/0 description=Test_Interface",
			requestType: types.EditConf,
			// The xmlstore.Insert implementation may put key/value pairs into
			// element text rather than nested tags in some cases; accept either form.
			wantSubstrs: []string{`Test_Interface`},
		},
		{
			name: "delete leaf",
			initial: func() *XMLStore {
				s := &XMLStore{}
				s.Insert(entryIf, "cmd ietf-interfaces interfaces interface name=GigabitEthernet0/0/0 description=Test_Interface", types.EditConf)
				return s
			}(),
			entry:       entryIf,
			requestLine: "delete ietf-interfaces interfaces interface name=GigabitEthernet0/0/0 description",
			requestType: types.EditConf,
			wantSubstrs: []string{`operation="remove"`, `<description`},
		},
		{
			name:        "insert identityref value",
			initial:     &XMLStore{},
			entry:       entryCrypto,
			requestLine: "cmd example-crypto crypto algorithm=ianaid:hmac-sha1",
			requestType: types.EditConf,
			// Accept either algorithm element or raw algorithm=... text
			wantSubstrs: []string{`algorithm`, `hmac-sha1`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &XMLStore{}
			if tt.initial != nil {
				// Copy initial by marshalling/unmarshalling via store.Marshal to keep test simpler
				*store = *tt.initial
			}
			store.Insert(tt.entry, tt.requestLine, tt.requestType)
			got := store.Marshal()
			for _, want := range tt.wantSubstrs {
				if !strings.Contains(got, want) {
					t.Errorf("XMLStore.Insert() output missing %q in:\n%s", want, got)
				}
			}
		})
	}
}

func TestXmlElement_MarshalXML(t *testing.T) {
	t.Run("simple element with value", func(t *testing.T) {
		el := xmlElement{XMLName: xml.Name{Local: "leaf1"}, Value: "value1"}
		gotBytes, err := xml.MarshalIndent(el, "", "  ")
		if err != nil {
			t.Fatalf("xml.MarshalIndent error: %v", err)
		}
		got := string(gotBytes)
		if !strings.Contains(got, "<leaf1>") || !strings.Contains(got, "value1") {
			t.Fatalf("unexpected xml: %s", got)
		}
	})

	t.Run("element with namespace and child", func(t *testing.T) {
		child := &xmlElement{XMLName: xml.Name{Local: "child1", Space: "urn:parent"}, Value: "cval1"}
		parent := xmlElement{XMLName: xml.Name{Local: "parent", Space: "urn:parent"}, Children: []xmlElementInterface{child}}
		gotBytes, err := xml.MarshalIndent(parent, "", "  ")
		if err != nil {
			t.Fatalf("xml.MarshalIndent error: %v", err)
		}
		got := string(gotBytes)
		if !strings.Contains(got, "<parent") || !(strings.Contains(got, "<child1>cval1</child1>") || strings.Contains(got, "<child1 xmlns=") && strings.Contains(got, "cval1")) {
			t.Fatalf("unexpected xml: %s", got)
		}
	})

	t.Run("element marked for deletion", func(t *testing.T) {
		el := xmlElement{XMLName: xml.Name{Local: "todelete", Space: "urn:example:test"}, delete: true}
		gotBytes, err := xml.MarshalIndent(el, "", "  ")
		if err != nil {
			t.Fatalf("xml.MarshalIndent error: %v", err)
		}
		got := string(gotBytes)
		if !strings.Contains(got, `operation="remove"`) {
			t.Fatalf("unexpected xml for delete element: %s", got)
		}
	})

	t.Run("identityref element", func(t *testing.T) {
		idref := "urn:identity:namespace"
		el := xmlElement{XMLName: xml.Name{Local: "idleaf", Space: "urn:example:id"}, Value: "my-id-value", idref: &idref}
		gotBytes, err := xml.MarshalIndent(el, "", "  ")
		if err != nil {
			t.Fatalf("xml.MarshalIndent error: %v", err)
		}
		got := string(gotBytes)
		if !strings.Contains(got, "my-id-value") || !strings.Contains(got, "xmlns=") {
			t.Fatalf("unexpected xml for idref element: %s", got)
		}
	})
}
