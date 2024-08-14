package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sendmail-to-matrix",
	Short: "Read an email message from STDIN and forward it to a Matrix room",
	RunE:  forwardCmd.RunE,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flags
	pf := rootCmd.PersistentFlags()
	pf.StringP("config-dir", "c", DefaultConfigDir, "Path to config directory")
	pf.Bool("no-encrypt", false, "Do not use encryption when sending messages")
	pf.String("config-file", "", "(deprecated) Path to config file")
	pf.String("room", "", "Matrix Room ID")
	pf.String("server", "", "Matrix server")
	pf.String("preface", "", "Preface the matrix message with text")
	pf.String("epilogue", "", "Append the matrix message with text")
	pf.String("template", "", "Template string used to render the Message")
	pf.StringArray("skip", []string{}, "Regex pattern that will skip sending a message if it matches")
	pf.String("token", "", "Token used to send non-encrypted messages")
	pf.String("db-pass", "", "Password used to secure the state database for encrypted messaging")
	viperConf.BindPFlags(pf)
}
