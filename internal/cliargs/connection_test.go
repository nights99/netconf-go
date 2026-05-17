package cliargs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func TestLoadDefaults(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	AddFlags(fs)

	cfg, err := Load(fs, t.TempDir())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Address != "localhost" {
		t.Fatalf("Address = %q, want %q", cfg.Address, "localhost")
	}
	if cfg.Port != 22 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 22)
	}
	if cfg.Debug != "info" {
		t.Fatalf("Debug = %q, want %q", cfg.Debug, "info")
	}
}

func TestLoadHostConfig(t *testing.T) {
	dir := t.TempDir()
	writeHostsFile(t, dir, "lab:\n  address: router.example\n  port: 830\n  user: admin\n  password: secret\n")

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	AddFlags(fs)
	if err := fs.Parse([]string{"--host", "lab"}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	cfg, err := Load(fs, dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Address != "router.example" {
		t.Fatalf("Address = %q, want %q", cfg.Address, "router.example")
	}
	if cfg.Port != 830 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 830)
	}
	if cfg.User != "admin" {
		t.Fatalf("User = %q, want %q", cfg.User, "admin")
	}
	if cfg.Password != "secret" {
		t.Fatalf("Password = %q, want %q", cfg.Password, "secret")
	}
}

func TestLoadFlagsOverrideHostConfig(t *testing.T) {
	dir := t.TempDir()
	writeHostsFile(t, dir, "lab:\n  address: router.example\n  port: 830\n  user: admin\n")

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	AddFlags(fs)
	if err := fs.Parse([]string{"--host", "lab", "--address", "override.example", "--user", "operator"}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	cfg, err := Load(fs, dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Address != "override.example" {
		t.Fatalf("Address = %q, want %q", cfg.Address, "override.example")
	}
	if cfg.User != "operator" {
		t.Fatalf("User = %q, want %q", cfg.User, "operator")
	}
	if cfg.Port != 830 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 830)
	}
}

func writeHostsFile(t *testing.T, dir string, contents string) {
	t.Helper()

	path := filepath.Join(dir, "hosts.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
