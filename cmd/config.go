package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tnyeanderson/sendmail-to-matrix/pkg"
)

const (
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

type config struct {
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

	ignoreConfigFileErrors bool
	skipsRegexp            []*regexp.Regexp
}

func (c *config) writeTo(w io.Writer) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func (c *config) init() error {
	v, err := newViper(rootCmd.PersistentFlags())
	if err != nil {
		return err
	}
	return c.fromViper(v)
}

func (c *config) fromViper(v *viper.Viper) error {
	configFile := v.GetString(flagConfigFile)

	if configFile != "" {
		if err := readConfigFile(v, configFile); err != nil && !c.ignoreConfigFileErrors {
			return err
		}
	}

	if err := v.Unmarshal(c); err != nil {
		return err
	}

	if c.Template == "" {
		c.Template = pkg.DefaultMessageTemplate
	}

	for _, skip := range c.Skip {
		r, err := regexp.Compile(skip)
		if err != nil {
			return err
		}
		c.skipsRegexp = append(c.skipsRegexp, r)
	}

	return nil
}

func newViper(f *pflag.FlagSet) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("json")
	if err := v.BindPFlags(f); err != nil {
		return nil, err
	}
	v.SetEnvPrefix("stm")
	v.AutomaticEnv()
	v.SetDefault(flagConfigDir, getDefaultConfigDir())
	v.SetDefault(flagConfigFile, getConfigFilePath(v))
	return v, nil
}

func getDefaultConfigDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "sendmail-to-matrix")
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
