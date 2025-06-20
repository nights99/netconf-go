module netconf-go

go 1.23.0

toolchain go1.24.1

require (
	github.com/chzyer/readline v1.5.1
	github.com/gobwas/ws v1.4.0
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/openconfig/goyang v1.4.5
	github.com/peterh/liner v1.2.2
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.38.0
	golang.org/x/sys v0.33.0 // indirect
	nhooyr.io/websocket v1.8.17
)

require github.com/Juniper/go-netconf v0.3.1

require (
	github.com/nemith/netconf v0.0.2
	github.com/spf13/pflag v1.0.6
	github.com/spf13/viper v1.20.1
	github.com/ziutek/telnet v0.1.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/text v0.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// go mod edit -replace github.com/openconfig/goyang=github.com/nights99/goyang@dynamic_read
// replace github.com/openconfig/goyang => ./goyang/

// replace github.com/openconfig/goyang => ./src/github.com/openconfig/goyang/
replace github.com/openconfig/goyang => github.com/nights99/goyang v0.2.5-0.20241208122904-7fab041cb7ce

// replace github.com/peterh/liner => ./src/github.com/peterh/liner

// go mod edit -replace github.com/Juniper/go-netconf=github.com/nights99/go-netconf@master
// replace github.com/Juniper/go-netconf => ./src/github.com/Juniper/go-netconf/
// replace github.com/Juniper/go-netconf => ./go-netconf/
// replace github.com/nemith/netconf => ./go-netconf-v2/

// replace github.com/Juniper/go-netconf => github.com/nights99/go-netconf v0.1.2-0.20220723134019-7f4f80450f34
