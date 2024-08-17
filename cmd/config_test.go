package cmd

import (
	"log"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestGetConfigFilePath(t *testing.T) {
	var expected string
	v := viper.New()
	viperConfInit(v, pflag.NewFlagSet("tmp", pflag.PanicOnError))

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

func TestGetConfigFromEnv(t *testing.T) {
	expected := "mytoken"
	t.Setenv("STM_TOKEN", expected)
	v := viper.New()
	viperConfInit(v, pflag.NewFlagSet("tmp", pflag.PanicOnError))

	// Don't read config from filesystem
	v.Set(flagConfigDir, "")
	v.Set(flagConfigFile, "")

	c, err := getConfig(v)
	if err != nil {
		log.Fatal(err)
	}
	if got := c.Token; got != expected {
		t.Fatalf("expected: %s, got %s", expected, got)
	}
}
