module netconf-go

go 1.15

require (
	github.com/Juniper/go-netconf v0.1.2-0.20201208192613-a527e68d123f
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/gobwas/ws v1.0.4
	github.com/openconfig/goyang v0.2.4
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	nhooyr.io/websocket v1.8.6
)

replace github.com/Juniper/go-netconf => /home/jon/go/src/github.com/Juniper/go-netconf/
