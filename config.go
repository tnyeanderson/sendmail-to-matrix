package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// I know, "this should be a struct!"
// But flags need to be parsed first, applying conf file values if not set by flags
// Check the parseConfigFile and validateConfigOrDie functions first
// Then try to convince me to make it a struct for no benefit :)
type Config map[string]string

var config Config

func parseFlags() {
	// Parse the flags
	version := flag.Bool("version", false, "Print the application version and exit")
	configFile := flag.String("config-file", "", "Path to config file")
	server := flag.String("server", "", "Matrix homeserver url")
	token := flag.String("token", "", "Matrix account access token")
	room := flag.String("room", "", "Matrix Room ID")
	preface := flag.String("preface", "", "Preface the matrix message with arbitrary text (optional)")
	flag.Parse()

	// If the version flag is set, print the version and exit
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	// Put them in the config
	config["configFile"] = *configFile
	config["server"] = *server
	config["token"] = *token
	config["room"] = *room
	config["preface"] = *preface
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

	// Save to file
	err = ioutil.WriteFile(configFile, b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// Notify user
	fmt.Println("")
	fmt.Printf("Saved config to: %s\n", configFile)
}

func getConfig() {
	config = make(Config)

	// Parse command line arguments. These override config file values.
	// We need it parsed before for the --config-file flag
	parseFlags()

	// Parse config file (if set)
	if config["configFile"] != "" {
		parseConfigFile(config["configFile"])
	}
}

func parseConfigFile(path string) {
	var confFile Config

	// Read config file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Unmarshall the contents
	err = json.Unmarshal(content, &confFile)
	if err != nil {
		log.Fatal("Error trying to parse JSON file: ", err)
	}

	// Set the values only if they weren't already set by flags
	keys := []string{"server", "token", "room", "preface"}
	for _, key := range keys {
		if config[key] == "" {
			config[key] = confFile[key]
		}
	}
}

func validateConfigOrDie() {
	var missing []string
	required := []string{"server", "token", "room"}
	for _, key := range required {
		if config[key] == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		log.Fatal("Missing required parameters: ", strings.Join(missing, ", "))
	}
}
