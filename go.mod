module netconf-go

go 1.17

require (
	github.com/Juniper/go-netconf v0.1.2-0.20201208192613-a527e68d123f
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/gobwas/ws v1.0.2
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/openconfig/goyang v0.2.4
	github.com/peterh/liner v1.2.1
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/arch v0.0.0-20210427114910-4d4a2a2eb4cf
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887
	nhooyr.io/websocket v1.8.6
)

require (
	github.com/chzyer/logex v1.1.10
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1
	github.com/nbutton23/zxcvbn-go v0.0.0-20210217022336-fa2cb2858354
	golang.org/x/text v0.3.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gobwas/httphead v0.0.0-20180130184737-2c6c146eadee // indirect
	github.com/gobwas/pool v0.2.0 // indirect
	github.com/golang/protobuf v1.4.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

// replace github.com/Juniper/go-netconf => ./src/github.com/Juniper/go-netconf/
// replace github.com/Juniper/go-netconf => ./go-netconf/
replace github.com/Juniper/go-netconf => github.com/nights99/go-netconf v0.1.2-0.20210724124515-822f771b087f

// Just do "go get github.com/nights99/go-netconf@chunk_read" and it fills this in?
// replace github.com/openconfig/goyang => github.com/nights99/goyang v0.2.5-0.20210118142943-720a812d72ab

// replace github.com/openconfig/goyang => ./goyang/
replace github.com/openconfig/goyang => ./src/github.com/openconfig/goyang/
