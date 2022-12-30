package main

import (
	"flag"
	"fmt"
)

var usage = `
Read an email message from STDIN and forward it to a Matrix room

USAGE:

  sendmail-to-matrix [OPTIONS...]

CONFIGURATION:

  A server, token, and room must be set either using a config file or via command-line parameters.

  Config file format (JSON):
  {
    "server": "https://matrix.example.org",
    "token": "<access token>"
    "room": "!roomid:homeservername"
    "preface": "Preface to message"
  }

OPTIONS:

`

var printUsage = func() {
	fmt.Fprint(flag.CommandLine.Output(), usage)
	flag.PrintDefaults()
	fmt.Fprint(flag.CommandLine.Output(), "\n")
}
