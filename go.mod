module netconf-go

go 1.17

require (
	github.com/Juniper/go-netconf v0.1.2-0.20201208192613-a527e68d123f
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/gin-gonic/gin v1.7.7
	github.com/gobwas/ws v1.1.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/openconfig/goyang v0.3.2
	github.com/peterh/liner v1.2.2
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9
	nhooyr.io/websocket v1.8.7
)

require (
	github.com/chzyer/logex v1.2.0
	github.com/chzyer/test v0.0.0-20210722231415-061457976a23
	github.com/nbutton23/zxcvbn-go v0.0.0-20210217022336-fa2cb2858354
	golang.org/x/text v0.3.7
)

require (
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/klauspost/compress v1.14.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/openconfig/gnmi v0.0.0-20210914185457-51254b657b7d // indirect
	github.com/openconfig/grpctunnel v0.0.0-20211112160204-16444a7ba84c // indirect
	github.com/openconfig/ygot v0.13.2 // indirect
	github.com/pborman/getopt v0.0.0-20190409184431-ee0cd42419d3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/stretchr/objx v0.1.0 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20220118154757-00ab72f36ad5 // indirect
	google.golang.org/grpc v1.43.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// replace github.com/Juniper/go-netconf => ./src/github.com/Juniper/go-netconf/
// replace github.com/Juniper/go-netconf => ./go-netconf/
replace github.com/Juniper/go-netconf => github.com/nights99/go-netconf v0.1.2-0.20210724124515-822f771b087f

// Just do "go get github.com/nights99/go-netconf@chunk_read" and it fills this in?
// replace github.com/openconfig/goyang => github.com/nights99/goyang v0.2.5-0.20210118142943-720a812d72ab

// replace github.com/openconfig/goyang => ./goyang/
// replace github.com/openconfig/goyang => ./src/github.com/openconfig/goyang/

// replace github.com/peterh/liner => ./src/github.com/peterh/liner
