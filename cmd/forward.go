package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tnyeanderson/sendmail-to-matrix/pkg"
)

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Read an email message from stdin and forward it to a Matrix room",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getConfig()
		if err != nil {
			return err
		}

		message, err := buildMessage(os.Stdin, c.Template, c.Preface, c.Epilogue)
		if err != nil {
			return err
		}

		if !pkg.FilterMessage(string(message), c.skipsRegexp) {
			fmt.Println("forwarding skipped due to filters")
		}

		if c.EncryptionDisabled {
			return forwardWithoutEncryption(c, c.Room, message)
		}
		return forward(c, c.Room, message)
	},
}

func buildMessage(r io.Reader, template, preface, epilogue string) ([]byte, error) {
	m, err := pkg.NewMessage(r)
	if err != nil {
		return nil, err
	}
	m.Preface = preface
	m.Epilogue = epilogue
	return m.Render([]byte(template))
}

func forward(config *cliConfig, room string, message []byte) error {
	dbPath := filepath.Join(config.ConfigDir, "stm.db")
	ctx := context.Background()
	logger := zerolog.New(os.Stderr)
	c, err := pkg.NewEncryptedClient(ctx, dbPath, config.DatabasePassword, logger)
	if err != nil {
		return err
	}
	return c.SendMessage(ctx, room, message)
}

func forwardWithoutEncryption(config *cliConfig, room string, message []byte) error {
	ctx := context.Background()
	c, err := pkg.NewUnencryptedClient(config.Server, config.Token)
	if err != nil {
		return err
	}
	return c.SendMessage(ctx, room, message)
}

func init() {
	rootCmd.AddCommand(forwardCmd)
}
