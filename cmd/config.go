package cmd

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
	"github.com/tnyeanderson/sendmail-to-matrix/pkg"
)

const DefaultConfigDir = "/etc/sendmail-to-matrix"

var viperConf = viper.New()

type cliConfig struct {
	ConfigDir          string `mapstructure:"config-dir"`
	DatabasePassword   string `mapstructure:"db-pass"`
	EncryptionDisabled bool   `mapstructure:"no-encrypt"`
	Epilogue           string
	Preface            string
	Room               string
	Server             string
	Skips              []string
	Template           string
	Token              string

	skipsRegexp []*regexp.Regexp
}

func getConfig() (*cliConfig, error) {
	if err := readViperConfigFromFile(); err != nil {
		return nil, err
	}

	c := &cliConfig{}
	if err := viperConf.Unmarshal(c); err != nil {
		return nil, err
	}

	for _, skip := range c.Skips {
		r, err := regexp.Compile(skip)
		if err != nil {
			return nil, err
		}
		c.skipsRegexp = append(c.skipsRegexp, r)
	}

	return c, nil
}

func readViperConfigFromFile() error {
	configDir := viperConf.GetString("config-dir")
	configFile := viperConf.GetString("config-file")

	if configDir == "" {
		// Skip trying to read config file if configDir is explicitly empty
		return nil
	}

	configFile := viperConf.GetString("config-file")
	if configFile == "" {
		configFile = filepath.Join(configDir, "config.json")
	}

	r, err := os.Open(configFile)
	if err != nil {
		return err
	}

	viperConf.SetConfigType("json")
	return viperConf.ReadConfig(r)
}

func init() {
	// Set default values
	viperConf.SetDefault("config-dir", DefaultConfigDir)
	viperConf.SetDefault("template", pkg.DefaultMessageTemplate)

	// This allows us to save the data in the Skips slice in the struct, but
	// provide multiple singular "--skip" flags.
	viperConf.RegisterAlias("skips", "skip")
}
