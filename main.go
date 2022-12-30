package main

import (
	"flag"
	"os"
)

var Version = "v0.0.1"

func main() {
	flag.Usage = printUsage
	getConfig()
	validateConfigOrDie()
	message := buildMessage(os.Stdin)
	sendMessage(message)
}
