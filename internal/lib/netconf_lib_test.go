package lib

import (
	"bytes"
	"encoding/xml"
	"netconf-go/internal/types"
	"netconf-go/internal/xmlstore"
	"reflect"
	"testing"

	netconf "github.com/nemith/netconf"
	"github.com/openconfig/goyang/pkg/yang"
)

func TestMain(m *testing.M) {
	m.Run()
}

func Test_listYang(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := listYang(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listYang() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_newNetconfRequest tests the newNetconfRequest function.
func Test_newNetconfRequest(t *testing.T) {
	type args struct {
		netconfEntry yang.Entry
		Path         []string
		value        string
		requestType  types.RequestType
		delete       bool
	}
	tests := []struct {
		name string
		args args
		want *netconfRequest
	}{
		{
			name: "empty path, empty value, EditConf, no delete",
			args: args{
				netconfEntry: yang.Entry{},
				Path:         []string{},
				value:        "",
				requestType:  types.EditConf,
				delete:       false,
			},
			want: &netconfRequest{
				ncEntry:     &yang.Entry{},
				NetConfPath: []netconfPathElement{},
				Value:       "",
				reqType:     types.EditConf,
			},
		},
		{
			name: "single element path, non-empty value, GetConf, no delete",
			args: args{
				netconfEntry: yang.Entry{Name: "interface"},
				Path:         []string{"interfaces", "interface[name=eth0]", "description"},
				value:        "My interface",
				requestType:  types.GetConf,
				delete:       false,
			},
			want: &netconfRequest{
				ncEntry: &yang.Entry{Name: "interface"},
				NetConfPath: []netconfPathElement{
					{name: "interfaces", value: nil, delete: false},
					{name: "interface[name=eth0]", value: nil, delete: false},
					{name: "description", value: nil, delete: false},
				},
				Value:   "My interface",
				reqType: types.GetConf,
			},
		},
		{
			name: "multi-element path, empty value, EditConf, delete",
			args: args{
				netconfEntry: yang.Entry{Name: "interface"},
				Path:         []string{"interfaces", "interface[name=eth0]", "mtu"},
				value:        "",
				requestType:  types.EditConf,
				delete:       true,
			},
			want: &netconfRequest{
				ncEntry: &yang.Entry{Name: "interface"},
				NetConfPath: []netconfPathElement{
					{name: "interfaces", value: nil, delete: false},
					{name: "interface[name=eth0]", value: nil, delete: false},
					{name: "mtu", value: nil, delete: true},
				},
				Value:   "",
				reqType: types.EditConf,
			},
		},
		{
			name: "RpcOp request type",
			args: args{
				netconfEntry: yang.Entry{Name: "get-schema"},
				Path:         []string{"get-schema", "identifier"},
				value:        "ietf-interfaces",
				requestType:  types.RpcOp,
				delete:       false,
			},
			want: &netconfRequest{
				ncEntry: &yang.Entry{Name: "get-schema"},
				NetConfPath: []netconfPathElement{
					{name: "get-schema", value: nil, delete: false},
					{name: "identifier", value: nil, delete: false},
				},
				Value:   "ietf-interfaces",
				reqType: types.RpcOp,
			},
		},
		{
			name: "path with key=value pair",
			args: args{
				netconfEntry: yang.Entry{Name: "system"},
				Path:         []string{"system", "services", "ssh", "port=22"},
				value:        "",
				requestType:  types.EditConf,
				delete:       false,
			},
			want: &netconfRequest{
				ncEntry: &yang.Entry{Name: "system"},
				NetConfPath: []netconfPathElement{
					{name: "system", value: nil, delete: false},
					{name: "services", value: nil, delete: false},
					{name: "ssh", value: nil, delete: false},
					{name: "port", value: func() *string { s := "22"; return &s }(), delete: false},
				},
				Value:   "",
				reqType: types.EditConf,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedRequest := tt.want
			// Since ncEntry is a pointer in netconfRequest, and we are comparing the Name field,
			// ensure that the wanted ncEntry is not nil if an args.netconfEntry.Name is provided.
			if tt.args.netconfEntry.Name != "" && expectedRequest.ncEntry == nil {
				expectedRequest.ncEntry = &yang.Entry{Name: tt.args.netconfEntry.Name}
			} else if tt.args.netconfEntry.Name != "" && expectedRequest.ncEntry != nil && expectedRequest.ncEntry.Name == "" {
				expectedRequest.ncEntry.Name = tt.args.netconfEntry.Name
			}


			got := newNetconfRequest(&tt.args.netconfEntry, tt.args.Path, tt.args.value, tt.args.requestType, tt.args.delete)

			if (got.ncEntry == nil && expectedRequest.ncEntry != nil) || (got.ncEntry != nil && expectedRequest.ncEntry == nil) {
				t.Errorf("newNetconfRequest() ncEntry nilness mismatch: got %v, want %v", got.ncEntry, expectedRequest.ncEntry)
			} else if got.ncEntry != nil && expectedRequest.ncEntry != nil && got.ncEntry.Name != expectedRequest.ncEntry.Name {
				t.Errorf("newNetconfRequest() ncEntry.Name = %q, want %q", got.ncEntry.Name, expectedRequest.ncEntry.Name)
			}

			if !reflect.DeepEqual(got.NetConfPath, expectedRequest.NetConfPath) {
				t.Errorf("newNetconfRequest() NetConfPath = %v, want %v", got.NetConfPath, expectedRequest.NetConfPath)
			}
			if got.Value != expectedRequest.Value {
				t.Errorf("newNetconfRequest() Value = %q, want %q", got.Value, expectedRequest.Value)
			}
			if got.reqType != expectedRequest.reqType {
				t.Errorf("newNetconfRequest() reqType = %v, want %v", got.reqType, expectedRequest.reqType)
			}
		})
	}
}

// Test_emitNestedXML tests the emitNestedXML function.
func Test_emitNestedXML(t *testing.T) {
	type args struct {
		paths   []netconfPathElement
		value   string
		reqType types.RequestType
	}
	tests := []struct {
		name string
		args args
		want string // Expected XML string
	}{
		{
			name: "simple path with value",
			args: args{
				paths: []netconfPathElement{
					{name: "urn:ietf:params:xml:ns:yang:ietf-interfaces@interfaces"},
					{name: "urn:ietf:params:xml:ns:yang:ietf-interfaces@interface"},
					{name: "urn:ietf:params:xml:ns:yang:ietf-interfaces@description"},
				},
				value:   "Core router interface",
				reqType: types.EditConf,
			},
			want: `<interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces"><interface xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces"><description xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">Core router interface</description></interface></interfaces>`,
		},
		{
			name: "path with delete operation and value for key",
			args: args{
				paths: []netconfPathElement{
					{name: "urn:ietf:params:xml:ns:yang:ietf-interfaces@interfaces"},
					{name: "urn:ietf:params:xml:ns:yang:ietf-interfaces@interface", delete: true},
					{name: "urn:ietf:params:xml:ns:yang:ietf-interfaces@name"},
				},
				value:   "eth1",
				reqType: types.EditConf,
			},
			want: `<interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces"><interface operation="remove" xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces"><name xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">eth1</name></interface></interfaces>`,
		},
		{
			name: "empty path, no value",
			args: args{
				paths:   []netconfPathElement{},
				value:   "",
				reqType: types.GetConf,
			},
			want: ``,
		},
		{
			name: "simple path with intermediate value",
			args: args{
				paths: []netconfPathElement{
					{name: "system"},
					{name: "login"},
					{name: "user", value: func() *string { s := "testuser"; return &s }()},
					{name: "class"},
				},
				value:   "superuser",
				reqType: types.EditConf,
			},
			want: `<system><login><user>testuser</user><class>superuser</class></login></system>`,
		},
		{
			name: "RpcOp type",
			args: args{
				paths: []netconfPathElement{
					{name: "urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring@get-schema"},
					{name: "urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring@identifier"},
				},
				value:   "ietf-interfaces",
				reqType: types.RpcOp,
			},
			want: `<get-schema xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring"><identifier xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring">ietf-interfaces</identifier></get-schema>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := xml.NewEncoder(&buf)
			if err := emitNestedXML(enc, tt.args.paths, tt.args.value, tt.args.reqType); err != nil {
				if !(len(tt.args.paths) == 0 && tt.want == "") { // Allow no error for empty path case
					t.Fatalf("emitNestedXML() returned an error: %v", err)
				}
			}
			if err := enc.Flush(); err != nil {
				t.Fatalf("enc.Flush() error = %v", err)
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("emitNestedXML() generated XML = %q, want %q", got, tt.want)
			}
		})
	}
}

// Test_netconfRequest_MarshalMethod tests the MarshalXML method of netconfRequest.
func Test_netconfRequest_MarshalMethod(t *testing.T) {
	type fields struct {
		ncEntry     yang.Entry
		NetConfPath []netconfPathElement
		Value       string
		reqType     types.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "EditConf - structure only (empty config)",
			fields: fields{
				ncEntry:     yang.Entry{Name: "top"}, // For MarshalXML, ncEntry.Name might be used.
				NetConfPath: []netconfPathElement{},  // store.Root will be empty
				Value:       "",
				reqType:     types.EditConf,
			},
			want: `<edit-config><target><candidate></candidate></target><config xmlns="urn:ietf:params:xml:ns:netconf:base:1.0"></config></edit-config>`,
		},
		{
			name: "GetConf with filter (empty store.Root)",
			fields: fields{
				ncEntry:     yang.Entry{Name: "interfaces"}, // ncEntry presence triggers <filter> tag
				NetConfPath: []netconfPathElement{},      // store.Root will be empty
				Value:       "",
				reqType:     types.GetConf,
			},
			want: `<get-config><source><running></running></source><filter type="subtree"></filter></get-config>`,
		},
		{
			name: "RpcOp (e.g. get-schema with children, assuming store populates)",
			fields: fields{
				ncEntry: yang.Entry{Name: "get-schema", // Root RPC tag
					// Mock parts of yang.Entry that Namespace() might use.
					// A real yang.Entry.Namespace() is complex.
					// This is a simplified representation for testing.
					Parent: &yang.Module{
						Namespace: &yang.Value{Name: "urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring"},
					},
				},
				// NetConfPath & Value would inform store.Root for RPC's children.
				NetConfPath: []netconfPathElement{{name: "identifier"}},
				Value:       "ietf-interfaces",
				reqType:     types.RpcOp,
			},
			// This 'want' assumes store.Root is populated to <identifier>ietf-interfaces</identifier>
			// AND ncEntry's namespace is correctly used for the root tag.
			want: `<get-schema xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring"><identifier>ietf-interfaces</identifier></get-schema>`,
		},
		{
			name: "Commit operation",
			fields: fields{
				ncEntry:     yang.Entry{Name: "commit"},
				NetConfPath: []netconfPathElement{},
				Value:       "",
				reqType:     types.Commit,
			},
			want: `<commit></commit>`,
		},
		{
			name: "Validate operation",
			fields: fields{
				ncEntry:     yang.Entry{Name: "validate"},
				NetConfPath: []netconfPathElement{},
				Value:       "",
				reqType:     types.Validate,
			},
			want: `<validate><source><candidate></candidate></source></validate>`,
		},
		{
			name: "GetOper with filter (empty store.Root)",
			fields: fields{
				ncEntry:     yang.Entry{Name: "interfaces-state"}, // ncEntry presence triggers <filter> tag
				NetConfPath: []netconfPathElement{},
				Value:       "",
				reqType:     types.GetOper,
			},
			want: `<get><filter type="subtree"></filter></get>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := xmlstore.New()

			// Simulate store population for the RpcOp test case that expects children
			if tt.name == "RpcOp (e.g. get-schema, assumes store.Root populates children)" {
				// This is a simplified way to simulate content in store.Root.
				// A more robust mock would involve deeper interaction with xmlstore or direct Root assignment.
				type identifier struct {
					XMLName xml.Name `xml:"identifier"` // No namespace here, assumes it's handled by parent or not needed for child
					Value   string   `xml:",chardata"`
				}
				mockStore.Root = &identifier{Value: "ietf-interfaces"}
			}

			nc := &netconfRequest{
				ncEntry:     &tt.fields.ncEntry,
				NetConfPath: tt.fields.NetConfPath,
				Value:       tt.fields.Value,
				reqType:     tt.fields.reqType,
				store:       mockStore,
			}

			gotBytes, err := xml.MarshalIndent(nc, "", "  ")
			if err != nil {
				t.Fatalf("xml.MarshalIndent() error = %v", err)
			}
			got := string(gotBytes)

			if got != tt.want {
				t.Errorf("netconfRequest.MarshalXML() got = \n%s\n\nwant = \n%s", got, tt.want)
			}
		})
	}
}

func Test_getYangModule(t *testing.T) {
	type args struct {
		s       *netconf.Session
		yangMod string
	}
	tests := []struct {
		name string
		args args
		want *yang.Module
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock session 's' if necessary, depending on GetYangModule's implementation
			// For now, assume it can be nil or a simple mock if it doesn't dereference s heavily
			if got := GetYangModule(tt.args.s, tt.args.yangMod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetYangModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendNetconfRequest(t *testing.T) {
	type args struct {
		s           *netconf.Session
		requestLine string
		requestType types.RequestType
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock session 's' if necessary
			got, got1 := SendNetconfRequest(tt.args.s, tt.args.requestLine, tt.args.requestType)
			if got != tt.want {
				t.Errorf("SendNetconfRequest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SendNetconfRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getSchemaList(t *testing.T) {
	type args struct {
		s *netconf.Session
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock session 's' if necessary
			if got := GetSchemaList(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSchemaList() = %v, want %v", got, tt.want)
			}
		})
	}
}
