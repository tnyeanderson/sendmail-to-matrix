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

const (
	DefaultConfigDir         = "/etc/sendmail-to-matrix"
	DefaultDeviceDisplayName = "sendmail-to-matrix"
	DefaultServer            = "https://matrix.org"
)

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
	ConfigDir          string   `json:"config-dir,omitempty" mapstructure:"config-dir,omitempty"`
	ConfigFile         string   `json:"config-file,omitempty" mapstructure:"config-file,omitempty"`
	DatabasePassword   string   `json:"db-pass,omitempty" mapstructure:"db-pass,omitempty"`
	EncryptionDisabled bool     `json:"no-encrypt,omitempty" mapstructure:"no-encrypt,omitempty"`
	Epilogue           string   `json:"epilogue,omitempty" mapstructure:",omitempty"`
	Preface            string   `json:"preface,omitempty" mapstructure:",omitempty"`
	Room               string   `json:"room,omitempty" mapstructure:",omitempty"`
	Server             string   `json:"server,omitempty" mapstructure:",omitempty"`
	Skip               []string `json:"skip,omitempty" mapstructure:",omitempty"`
	Template           string   `json:"template,omitempty" mapstructure:",omitempty"`
	Token              string   `json:"token,omitempty" mapstructure:",omitempty"`

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

func getConfig(v *viper.Viper, ignoreConfigFileErrors bool) (*cliConfig, error) {
	configFile := getConfigFilePath(v)
	v.Set(flagConfigFile, configFile)

	if configFile != "" {
		if err := readConfigFile(v, configFile); err != nil && !ignoreConfigFileErrors {
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

func readConfigFile(v *viper.Viper, path string) error {
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := v.ReadConfig(r); err != nil {
		return err
	}
	return nil
}

func viperConfInit(v *viper.Viper, f *pflag.FlagSet) {
	rootFlagsInit(f)
	v.SetConfigType("json")
	v.BindPFlags(f)
	v.SetEnvPrefix("stm")
	v.AutomaticEnv()
	v.SetDefault(flagConfigDir, DefaultConfigDir)
}

func init() {
	viperConf = viper.New()
	viperConfInit(viperConf, rootCmd.PersistentFlags())
}
