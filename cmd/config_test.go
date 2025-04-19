package cmd

import (
	"log"
	"testing"

	"github.com/spf13/pflag"
)

func TestGetConfigFilePath(t *testing.T) {
	var expected string
	v, err := newViper(pflag.NewFlagSet("tmp", pflag.PanicOnError))
	if err != nil {
		t.Fatalf("failed to get viper config")
	}

	// Both empty string
	expected = ""
	v.Set(flagConfigDir, "")
	v.Set(flagConfigFile, "")
	if got := getConfigFilePath(v); got != expected {
		t.Fatalf("expected: empty string, got %s", got)
	}

	// configDir set
	expected = "my/config/path/config.json"
	v.Set(flagConfigDir, "my/config/path/")
	if got := getConfigFilePath(v); got != expected {
		t.Fatalf("expected: %s, got %s", expected, got)
	}

	// configFile set
	expected = "config/file/path.json"
	v.Set(flagConfigFile, expected)
	if got := getConfigFilePath(v); got != expected {
		t.Fatalf("expected: %s, got %s", expected, got)
	}
}

func TestConfigFromViper_Env(t *testing.T) {
	expected := "mytoken"
	t.Setenv("STM_TOKEN", expected)
	f := pflag.NewFlagSet("tmp", pflag.PanicOnError)
	rootFlagsInit(f)
	v, err := newViper(f)
	if err != nil {
		t.Fatalf("failed to get viper config")
	}

	// Don't read config from filesystem
	v.Set(flagConfigDir, "")
	v.Set(flagConfigFile, "")

	c := &config{ignoreConfigFileErrors: true}
	if err := c.fromViper(v); err != nil {
		log.Fatal(err)
	}

	if got := c.Token; got != expected {
		t.Fatalf("expected: %s, got %s", expected, got)
	}
}
