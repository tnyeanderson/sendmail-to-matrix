package main

import (
	"flag"
	"os"
)

var Version = "v0.1.2"

func main() {
	if os.Args[1] == "generate-config" {
		generateConfig()
		return
	}
	flag.Usage = printUsage
	getConfig()
	validateConfigOrDie()
	message := buildMessage(os.Stdin)
	sendMessage(message)
}
