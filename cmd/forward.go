package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

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

		return forward(c)
	},
}

func forward(c *config) error {
	// Set up tee to buffer message as we read it, so we can send the original
	// MIME if all else fails.
	originalEmail := new(bytes.Buffer)
	t := io.TeeReader(os.Stdin, originalEmail)

	message, err := buildMessage(t, c.Preface, c.Epilogue)
	if err != nil {
		// Read any unread data into the originalEmail tee
		if _, err := io.ReadAll(t); err != nil {
			return err
		}
		// If we couldn't build the message, upload the full email as a file.
		text := fmt.Sprintf("ERROR: sendmail-to-matrix failed to parse the attached email due to the following error: %s", err.Error())
		attachments := []pkg.Attachment{
			{
				Content:     originalEmail.Bytes(),
				ContentType: "message/rfc822",
				FileName:    "message.eml",
			},
		}
		return sendToMatrix(c, c.RoomID, []byte(text), attachments)
	}

	text, err := message.Render([]byte(c.Template))
	if err != nil {
		return err
	}

	if !filterMessage(string(text), c.skipsRegexp) {
		fmt.Println("forwarding skipped due to filters")
	}

	return sendToMatrix(c, c.RoomID, text, message.Attachments)
}

func buildMessage(r io.Reader, preface, epilogue string) (*pkg.Message, error) {
	m, err := pkg.NewMessageFromEmail(r)
	if err != nil {
		return nil, err
	}
	m.Preface = preface
	m.Epilogue = epilogue
	return m, nil
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

func newClientFromConfig(c *config) (pkg.Client, error) {
	if c.EncryptionDisabled {
		return pkg.NewUnencryptedClient(c.Server, c.UserID, c.Token)
	}

	// Encrypted messaging
	databasePath := filepath.Join(c.ConfigDir, "stm.db")
	return pkg.NewEncryptedClient(c.UserID, databasePath, c.DatabasePassword)
}

func sendToMatrix(conf *config, room string, text []byte, attachments []pkg.Attachment) error {
	ctx := context.Background()
	client, err := newClientFromConfig(conf)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.SendText(ctx, room, text); err != nil {
		return err
	}

	if conf.IncludeAttachments {
		for _, a := range attachments {
			if err := client.SendAttachment(ctx, room, a); err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(forwardCmd)
}
