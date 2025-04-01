package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tnyeanderson/sendmail-to-matrix/pkg"
)

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Read an email message from stdin and forward it to a Matrix room",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := &config{}
		if err := c.init(); err != nil {
			return err
		}

		// Set up tee to buffer message as we read it, so we can send the original
		// MIME if all else fails.
		original := new(bytes.Buffer)
		t := io.TeeReader(os.Stdin, original)

		message, err := buildMessage(t, c.Template, c.Preface, c.Epilogue)
		if err != nil {
			preface := []byte("ERROR: sendmail-to-matrix couldn't parse the below MIME message:\n\n")
			message = append(preface, original.Bytes()...)
		}

		if !filterMessage(string(message), c.skipsRegexp) {
			fmt.Println("forwarding skipped due to filters")
		}

		if c.EncryptionDisabled {
			return forwardWithoutEncryption(c, c.Room, message)
		}
		return forward(c, c.Room, message)
	},
}

func buildMessage(r io.Reader, template, preface, epilogue string) ([]byte, error) {
	m, err := pkg.NewMessageFromEmail(r)
	if err != nil {
		return nil, err
	}
	m.Preface = preface
	m.Epilogue = epilogue
	return m.Render([]byte(template))
}

// filterMessage returns true if the message should be forwarded to matrix and
// false if the message should be skipped/ignored.
func filterMessage(message string, skips []*regexp.Regexp) bool {
	for _, r := range skips {
		if r.MatchString(message) {
			return false
		}
	}
	return true
}

func forward(config *config, room string, message []byte) error {
	dbPath := filepath.Join(config.ConfigDir, "stm.db")
	ctx := context.Background()
	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	client, err := pkg.NewEncryptedClient(ctx, dbPath, config.DatabasePassword, logger)
	if err != nil {
		return err
	}
	return client.SendMessage(ctx, room, message)
}

func forwardWithoutEncryption(config *config, room string, message []byte) error {
	ctx := context.Background()
	client, err := pkg.NewUnencryptedClient(config.Server, config.Token)
	if err != nil {
		return err
	}
	return client.SendMessage(ctx, room, message)
}

func init() {
	rootCmd.AddCommand(forwardCmd)
}
