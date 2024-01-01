module netconf-go

go 1.18

require (
	github.com/chzyer/readline v1.5.1
	github.com/gobwas/ws v1.2.1
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/openconfig/goyang v1.4.5
	github.com/peterh/liner v1.2.2
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
	golang.org/x/crypto v0.17.0
	golang.org/x/sys v0.15.0 // indirect
	nhooyr.io/websocket v1.8.10
)

require (
	github.com/Juniper/go-netconf v0.3.0
	// github.com/nemith/go-netconf/v2 v2.0.0-00010101000000-000000000000
	github.com/nemith/netconf v0.0.1
	github.com/ziutek/telnet v0.0.0-20180329124119-c3b780dc415b
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20231226003508-02704c960a9b // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// go mod edit -replace github.com/openconfig/goyang=github.com/nights99/goyang@dynamic_read
replace github.com/openconfig/goyang => ./goyang/

// replace github.com/openconfig/goyang => ./src/github.com/openconfig/goyang/
// replace github.com/openconfig/goyang => github.com/nights99/goyang v0.2.5-0.20230528130339-76fd486cbc28

// replace github.com/peterh/liner => ./src/github.com/peterh/liner

// go mod edit -replace github.com/Juniper/go-netconf=github.com/nights99/go-netconf@master
// replace github.com/Juniper/go-netconf => ./src/github.com/Juniper/go-netconf/
// replace github.com/Juniper/go-netconf => ./go-netconf/
// replace github.com/nemith/go-netconf/v2 => ./go-netconf-v2/
replace github.com/nemith/netconf => ./go-netconf-v2/

// replace github.com/Juniper/go-netconf => github.com/nights99/go-netconf v0.1.2-0.20220723134019-7f4f80450f34
