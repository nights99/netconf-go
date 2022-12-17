module netconf-go

go 1.18

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/gin-gonic/gin v1.7.7 // indirect
	github.com/gobwas/ws v1.1.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/openconfig/goyang v0.3.2
	github.com/peterh/liner v1.2.2
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20221010152910-d6f0a8c073c2
	golang.org/x/sys v0.0.0-20221013171732-95e765b1cc43 // indirect
	nhooyr.io/websocket v1.8.7
)

require (
	github.com/Juniper/go-netconf v0.3.0
	github.com/nemith/go-netconf/v2 v2.0.0-00010101000000-000000000000
	github.com/ziutek/telnet v0.0.0-20180329124119-c3b780dc415b
)

require (
	github.com/chzyer/logex v1.2.0 // indirect
	github.com/chzyer/test v0.0.0-20210722231415-061457976a23 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/klauspost/compress v1.14.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nemith/netconf v0.0.0-20221130162605-a32beb732855 // indirect
	github.com/openconfig/gnmi v0.0.0-20210914185457-51254b657b7d // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// go mod edit -replace github.com/openconfig/goyang=github.com/nights99/goyang@dynamic_read
replace github.com/openconfig/goyang => ./goyang/

// replace github.com/openconfig/goyang => ./src/github.com/openconfig/goyang/
// replace github.com/openconfig/goyang => github.com/nights99/goyang v0.2.5-0.20220723135300-f046d3f17ec9

// replace github.com/peterh/liner => ./src/github.com/peterh/liner

// go mod edit -replace github.com/Juniper/go-netconf=github.com/nights99/go-netconf@master
// replace github.com/Juniper/go-netconf => ./src/github.com/Juniper/go-netconf/
// replace github.com/Juniper/go-netconf => ./go-netconf/
replace github.com/nemith/go-netconf/v2 => ./go-netconf-v2/

// replace github.com/Juniper/go-netconf => github.com/nights99/go-netconf v0.1.2-0.20220723134019-7f4f80450f34
