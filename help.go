package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
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
    "token": "<access token>",
    "room": "!roomid:homeservername",
    "preface": "Preface to message",
		"skips": ["Subject: Hello", "you've won [0-9]+ dollars"]
  }

OPTIONS:

`

var printUsage = func() {
	fmt.Fprint(os.Stderr, usage)
	pflag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}
