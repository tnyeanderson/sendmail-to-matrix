package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "sendmail-to-matrix",
	Short: "Read an email message from STDIN and forward it to a Matrix room",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func rootFlags(f *pflag.FlagSet) {
	f.StringP(flagConfigDir, "c", DefaultConfigDir, "Path to config directory, set to explicit empty string to skip reading all config files")
	f.String(flagConfigFile, "", "Path to JSON config file, defaults to config.json in the --config-dir path")
	f.Bool(flagNoEncrypt, false, "Do not use encryption when sending messages")
	f.String(flagRoom, "", "Matrix Room ID")
	f.String(flagServer, "", "Matrix home server URI")
	f.String(flagPreface, "", "Preface the matrix message with a line of text")
	f.String(flagEpilogue, "", "Append the matrix message with a line of text")
	f.String(flagTemplate, "", "Alternative template string used to render the message")
	f.StringArray(flagSkip, []string{}, "Regex patterns that will skip sending a rendered message if any match")
	f.String(flagToken, "", "Token used to send non-encrypted messages")
	f.String(flagDatabasePassword, "", "Password used to secure the state database for encrypted messaging")
}
