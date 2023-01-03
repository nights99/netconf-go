package main

import (
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/Juniper/go-netconf/netconf"
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

func Test_newNetconfRequest(t *testing.T) {
	type args struct {
		netconfEntry yang.Entry
		Path         []string
		value        string
		requestType  requestType
		delete       bool
	}
	tests := []struct {
		name string
		args args
		want *netconfRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newNetconfRequest(&tt.args.netconfEntry, tt.args.Path, tt.args.value, tt.args.requestType, tt.args.delete); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newNetconfRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_emitNestedXML(t *testing.T) {
	type args struct {
		enc     *xml.Encoder
		paths   []netconfPathElement
		value   string
		reqType requestType
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitNestedXML(tt.args.enc, tt.args.paths, tt.args.value, tt.args.reqType)
		})
	}
}

func Test_netconfRequest_MarshalMethod(t *testing.T) {
	type fields struct {
		ncEntry     yang.Entry
		NetConfPath []netconfPathElement
		Value       string
		reqType     requestType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nc := &netconfRequest{
				ncEntry:     &tt.fields.ncEntry,
				NetConfPath: tt.fields.NetConfPath,
				Value:       tt.fields.Value,
				reqType:     tt.fields.reqType,
			}
			if got := nc.MarshalMethod(); got != tt.want {
				t.Errorf("netconfRequest.MarshalMethod() = %v, want %v", got, tt.want)
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
			if got := getYangModule(tt.args.s, tt.args.yangMod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getYangModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendNetconfRequest(t *testing.T) {
	type args struct {
		s           *netconf.Session
		requestLine string
		requestType requestType
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
			got, got1 := sendNetconfRequest(tt.args.s, tt.args.requestLine, tt.args.requestType)
			if got != tt.want {
				t.Errorf("sendNetconfRequest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("sendNetconfRequest() got1 = %v, want %v", got1, tt.want1)
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
			if got := getSchemaList(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSchemaList() = %v, want %v", got, tt.want)
			}
		})
	}
}
