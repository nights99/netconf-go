module netconf-go

go 1.15

require (
	github.com/Juniper/go-netconf v0.1.2-0.20201208192613-a527e68d123f
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/davrodpin/mole v1.0.1 // indirect
	github.com/go-delve/delve v1.6.0 // indirect
	github.com/gobwas/ws v1.0.2
	github.com/google/go-dap v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/openconfig/goyang v0.2.4
	github.com/peterh/liner v1.2.1
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3 // indirect
	go.starlark.net v0.0.0-20210429133630-0c63ff3779a6 // indirect
	golang.org/dl v0.0.0-20210423174834-f798e20c9ec1 // indirect
	golang.org/x/arch v0.0.0-20210427114910-4d4a2a2eb4cf // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
	nhooyr.io/websocket v1.8.6
)

replace github.com/Juniper/go-netconf => ./src/github.com/Juniper/go-netconf/

replace github.com/openconfig/goyang => github.com/nights99/goyang v0.2.5-0.20210118142943-720a812d72ab
