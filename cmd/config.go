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
	// Read from file
	configFile := viperConf.GetString("config-file")
	if configFile == "" {
		configFile = filepath.Join(viperConf.GetString("config-dir"), "config.json")
	}
	r, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	viperConf.SetConfigType("json")
	if err := viperConf.ReadConfig(r); err != nil {
		// Ignore if config file not found
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
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

func init() {
	viperConf.RegisterAlias("skips", "skip")
	viperConf.SetDefault("config-dir", DefaultConfigDir)
	viperConf.SetDefault("template", pkg.DefaultMessageTemplate)
}
