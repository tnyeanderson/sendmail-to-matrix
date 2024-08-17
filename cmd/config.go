package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const DefaultConfigDir = "/etc/sendmail-to-matrix"

const (
	flagConfigDir        = "config-dir"
	flagConfigFile       = "config-file"
	flagNoEncrypt        = "no-encrypt"
	flagRoom             = "room"
	flagServer           = "server"
	flagPreface          = "preface"
	flagEpilogue         = "epilogue"
	flagTemplate         = "template"
	flagSkip             = "skip"
	flagToken            = "token"
	flagDatabasePassword = "db-pass"
)

var viperConf *viper.Viper

type cliConfig struct {
	ConfigDir          string `mapstructure:"config-dir"`
	ConfigFile         string `mapstructure:"config-file"`
	DatabasePassword   string `mapstructure:"db-pass"`
	EncryptionDisabled bool   `mapstructure:"no-encrypt"`
	Epilogue           string
	Preface            string
	Room               string
	Server             string
	Skip               []string
	Template           string
	Token              string

	skipsRegexp []*regexp.Regexp
}

func (c *cliConfig) writeTo(w io.Writer) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func getConfig(v *viper.Viper) (*cliConfig, error) {
	configFile := getConfigFilePath(v)
	v.Set(flagConfigFile, configFile)

	if configFile != "" {
		r, err := os.Open(configFile)
		if err != nil {
			return nil, err
		}
		if err := v.ReadConfig(r); err != nil {
			return nil, err
		}
	}

	c := &cliConfig{}
	if err := v.Unmarshal(c); err != nil {
		return nil, err
	}

	for _, skip := range c.Skip {
		r, err := regexp.Compile(skip)
		if err != nil {
			return nil, err
		}
		c.skipsRegexp = append(c.skipsRegexp, r)
	}

	return c, nil
}

func getConfigFilePath(v *viper.Viper) string {
	configFile := v.GetString(flagConfigFile)
	if configFile != "" {
		return configFile
	}

	configDir := v.GetString(flagConfigDir)
	if configDir == "" {
		return ""
	}

	return filepath.Join(configDir, "config.json")
}

func viperConfInit(v *viper.Viper, f *pflag.FlagSet) {
	v.SetConfigType("json")
	rootFlags(f)
	v.BindPFlags(f)
	v.SetEnvPrefix("stm")
	v.AutomaticEnv()
	v.SetDefault(flagConfigDir, DefaultConfigDir)
}

func init() {
	viperConf = viper.New()
	viperConfInit(viperConf, rootCmd.PersistentFlags())
}
