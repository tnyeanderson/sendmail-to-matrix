package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

type Config struct {
	ConfigFile  string   `json:"configFile"`
	Server      string   `json:"server"`
	Token       string   `json:"token"`
	Room        string   `json:"room"`
	Preface     string   `json:"preface"`
	Skips       []string `json:"skips"`
	skipsRegexp []*regexp.Regexp
}

func initConfig() (*Config, error) {
	pflag.Usage = printUsage

	// Flags
	fromFlags := parseFlags()

	// Config file
	fromFile := &Config{}
	if f := fromFlags.ConfigFile; f != "" {
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(content, fromFile)
		if err != nil {
			return nil, err
		}
	}

	// Merge
	c := mergeConfigs(*fromFlags, *fromFile)

	// Validate
	if err := validateConfig(c); err != nil {
		return nil, err
	}

	return c, nil
}

func parseFlags() *Config {
	c := &Config{}

	// Parse the flags
	pflag.StringVar(&(c.ConfigFile), "config-file", "", "Path to config file")
	pflag.StringVar(&(c.Server), "server", "", "Matrix homeserver url")
	pflag.StringVar(&(c.Token), "token", "", "Matrix account access token")
	pflag.StringVar(&(c.Room), "room", "", "Matrix Room ID")
	pflag.StringVar(&(c.Preface), "preface", "", "Preface the matrix message with arbitrary text (optional)")
	pflag.StringArrayVar(&(c.Skips), "skip", []string{}, "Regex pattern that will skip sending a message if it matches (optional)")
	version := pflag.Bool("version", false, "Print the application version and exit")
	pflag.Parse()

	// If the version flag is set, print the version and exit
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	return c
}

func mergeConfigs(fromFlags Config, fromFile Config) *Config {
	// Flags override config file values
	c := &fromFile
	if v := fromFlags.ConfigFile; v != "" {
		c.ConfigFile = v
	}
	if v := fromFlags.Server; v != "" {
		c.Server = v
	}
	if v := fromFlags.Token; v != "" {
		c.Token = v
	}
	if v := fromFlags.Room; v != "" {
		c.Room = v
	}
	if v := fromFlags.Preface; v != "" {
		c.Preface = v
	}
	// Skips from flags and config file are merged
	for _, skip := range fromFlags.Skips {
		c.Skips = append(c.Skips, skip)
	}
	return c
}

func validateConfig(c *Config) error {
	var missing []string
	if c.Server == "" {
		missing = append(missing, "server")
	}
	if c.Token == "" {
		missing = append(missing, "token")
	}
	if c.Room == "" {
		missing = append(missing, "room")
	}
	if len(missing) > 0 {
		return fmt.Errorf("Missing required parameters: %s", strings.Join(missing, ", "))
	}
	for _, skip := range c.Skips {
		r, err := regexp.Compile(skip)
		if err != nil {
			return err
		}
		c.skipsRegexp = append(c.skipsRegexp, r)
	}
	return nil
}

func generateConfig() {
	// Prompt user
	configFile := ask("Output path for generated config:")
	server := ask("Matrix home server (ex. https://matrix.org):")
	user := ask("Matrix username:")
	password := ask("Matrix password:")
	room := ask("Matrix room:")
	preface := ask("Message preface:")

	// Fetch access_token
	token, err := getToken(server, user, password)
	if err != nil {
		log.Fatal(err)
	}

	// Generate json
	data := map[string]string{
		"server":  server,
		"token":   token,
		"room":    room,
		"preface": preface,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Print the result
	fmt.Printf("\n%s\n\n", string(b))

	// Save to file
	err = ioutil.WriteFile(configFile, b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// Notify user
	fmt.Println("")
	fmt.Printf("Saved config to: %s\n", configFile)
}
