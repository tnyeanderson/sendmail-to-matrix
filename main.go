package main

import (
	"log"
	"os"
)

var Version = "v0.1.2"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "generate-config" {
		generateConfig()
		return
	}

	config, err := initConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	message := buildMessage(os.Stdin, config.Preface)
	if filterMessage(message, config.skipsRegexp) {
		sendMessage(config, message)
	}
}
