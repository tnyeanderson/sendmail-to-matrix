package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tnyeanderson/sendmail-to-matrix/pkg"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	_ "github.com/glebarez/go-sqlite"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive configuration utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := &config{ignoreConfigFileErrors: true}
		if err := c.init(); err != nil {
			return err
		}

		return setup(c)
	},
}

func setup(c *config) error {
	if err := os.MkdirAll(filepath.Dir(c.ConfigFile), 0750); err != nil {
		return err
	}

	f, err := os.OpenFile(c.ConfigFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if c.UserID == "" {
		c.UserID = ask("Matrix user ID", "")
	}

	userID := id.UserID(c.UserID)
	if _, _, err := userID.ParseAndValidate(); err != nil {
		return err
	}

	password := ask("Matrix password (not saved)", "")

	if c.Server == "" {
		c.Server = ask("Matrix home server", "https://"+userID.Homeserver())
	}

	homeserver, err := parseHomeserver(c.Server)
	if err != nil {
		return err
	}
	c.Server = homeserver

	if c.RoomID == "" {
		c.RoomID = ask("Matrix room ID", "")
	}

	// Unencrypted messaging
	if c.EncryptionDisabled {
		token, err := fetchToken(c.Server, userID.Localpart(), password)
		if err != nil {
			return err
		}
		c.Token = token
		if err := c.writeTo(f); err != nil {
			return err
		}
		fmt.Printf("\nSaved config to: %s\n", c.ConfigFile)
		return nil
	}

	// Encrypted messaging
	recoveryCode := ask("Recovery code (not saved, used for device verification)", "")
	deviceName := ask("Device display name", DefaultDeviceDisplayName)

	if c.DatabasePassword == "" {
		c.DatabasePassword = ask("Database encryption passphrase", "")
	}

	if err := c.writeTo(f); err != nil {
		return err
	}

	databasePath := filepath.Join(c.ConfigDir, "stm.db")

	_, err = pkg.SetupNewEncryptedClientUsingRecoveryKey(
		context.Background(),
		c.Server, c.UserID, password, recoveryCode,
		databasePath, c.DatabasePassword, deviceName,
	)

	if err != nil {
		return err
	}

	return nil
}

// fetchToken returns a simple access token for unencrypted messaging.
func fetchToken(server, username, password string) (string, error) {
	userID := id.NewUserID(username, server)
	client, err := mautrix.NewClient(server, userID, "")
	if err != nil {
		return "", err
	}

	res, err := client.Login(context.Background(), &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: userID.Localpart(),
		},
		Password: password,
	})
	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

func parseHomeserver(h string) (string, error) {
	// Remove any protocol prefix, if present
	if i := strings.LastIndex(h, "//"); i > -1 {
		h = h[i+2:]
	}

	// Ensure it is a valid URL (add the protocol prefix back)
	serverURL, err := url.Parse("//" + h)
	if err != nil {
		return "", err
	}

	return "https://" + serverURL.Hostname(), nil
}

// ask prompts the user for input and returns that input as a string.  If the
// user does not enter a value, defaultValue is returned. For convenience, it
// panics if it can't read from stdin.
func ask(prompt, defaultValue string) string {
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]: ", prompt, defaultValue)
	} else {
		prompt = prompt + ": "
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic("failed to read from stdin")
	}
	answer := scanner.Text()
	if answer == "" {
		return defaultValue
	}
	return answer
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
