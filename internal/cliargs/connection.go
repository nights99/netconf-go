package cliargs

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	flagAddress  = "address"
	flagDebug    = "debug"
	flagHost     = "host"
	flagPassword = "password"
	flagPort     = "port"
	flagTelnet   = "telnet"
	flagUser     = "user"
)

// ConnectionConfig holds reusable connection-related CLI arguments.
type ConnectionConfig struct {
	Address  string
	Debug    string
	Host     string
	Password string
	Port     int
	Telnet   bool
	User     string
}

// AddFlags registers the common connection flags on the provided flag set.
func AddFlags(fs *pflag.FlagSet) {
	fs.Int(flagPort, 22, "Port number to connect to")
	fs.String(flagAddress, "localhost", "Address or host to connect to")
	fs.String(flagHost, "", "Hostname referring to hosts.yaml entry")
	fs.BoolP(flagTelnet, "t", false, "Use telnet to connect")
	fs.String(flagDebug, log.InfoLevel.String(), "debug level")
	fs.String(flagUser, "", "Username")
	fs.String(flagPassword, "", "Password")
}

// Load resolves common connection flags, merging an optional hosts.yaml entry
// with explicit CLI overrides.
func Load(fs *pflag.FlagSet, configPaths ...string) (ConnectionConfig, error) {
	hosts := viper.New()
	hosts.SetConfigName("hosts")
	if len(configPaths) == 0 {
		configPaths = []string{"."}
	}
	for _, path := range configPaths {
		hosts.AddConfigPath(path)
	}

	if err := hosts.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return ConnectionConfig{}, fmt.Errorf("read hosts config: %w", err)
		}
	}

	source := viper.New()
	setDefaults(source)

	host, err := fs.GetString(flagHost)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("read host flag: %w", err)
	}

	if host != "" {
		hostConfig := hosts.Sub(host)
		if hostConfig == nil {
			return ConnectionConfig{}, fmt.Errorf("host configuration %q not found", host)
		}
		source = hostConfig
		setDefaults(source)
	}

	if err := source.BindPFlags(fs); err != nil {
		return ConnectionConfig{}, fmt.Errorf("bind flags: %w", err)
	}

	return ConnectionConfig{
		Address:  source.GetString(flagAddress),
		Debug:    source.GetString(flagDebug),
		Host:     host,
		Password: source.GetString(flagPassword),
		Port:     source.GetInt(flagPort),
		Telnet:   source.GetBool(flagTelnet),
		User:     source.GetString(flagUser),
	}, nil
}

func setDefaults(source *viper.Viper) {
	source.SetDefault(flagAddress, "localhost")
	source.SetDefault(flagDebug, log.InfoLevel.String())
	source.SetDefault(flagPort, 22)
}
