package main

import (
	"flag"
	"os"
)

var Version = "v0.1.2"

func main() {
	flag.Usage = printUsage
	getConfig()
	validateConfigOrDie()
	message := buildMessage(os.Stdin)
	sendMessage(message)
}
