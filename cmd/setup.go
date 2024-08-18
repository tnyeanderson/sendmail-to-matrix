package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tnyeanderson/sendmail-to-matrix/pkg"

	_ "github.com/mattn/go-sqlite3"
	_ "go.mau.fi/util/dbutil/litestream"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive configuration utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getConfig(viperConf, true)
		if err != nil {
			return err
		}

		if c.EncryptionDisabled {
			return setupWithoutEncryption(c)
		}

		return setup(c)
	},
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

func setup(config *cliConfig) error {
	if err := os.MkdirAll(config.ConfigDir, 0750); err != nil {
		return err
	}

	if cfDir := filepath.Dir(config.ConfigFile); config.ConfigDir != cfDir {
		if err := os.MkdirAll(cfDir, 0750); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(config.ConfigFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if config.Server == "" {
		config.Server = ask("Matrix home server", DefaultServer)
	}

	user := ask("Matrix username", "")
	password := ask("Matrix password (not saved)", "")
	recoveryCode := ask("Recovery code (not saved, used for device verification)", "")
	deviceName := ask("Device display name", "sendmail-to-matrix")

	if config.DatabasePassword == "" {
		config.DatabasePassword = ask("Database encryption passphrase", DefaultDeviceDisplayName)
	}

	if err := config.writeTo(f); err != nil {
		return err
	}

	dbPath := filepath.Join(config.ConfigDir, "stm.db")
	ctx := context.Background()
	logger := zerolog.New(os.Stderr)
	client, err := pkg.NewEncryptedClient(ctx, dbPath, config.DatabasePassword, logger)
	if err != nil {
		return err
	}
	return client.LoginAndVerify(ctx, config.Server, user, password, recoveryCode, deviceName)
}

func setupWithoutEncryption(config *cliConfig) error {
	if err := os.MkdirAll(filepath.Dir(config.ConfigFile), 0750); err != nil {
		return err
	}

	f, err := os.OpenFile(config.ConfigFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if config.Server == "" {
		config.Server = ask("Matrix home server", DefaultServer)
	}

	if config.Token == "" {
		user := ask("Matrix username", "")
		password := ask("Matrix password (not saved)", "")
		token, err := pkg.GetToken(config.Server, user, password)
		if err != nil {
			return err
		}
		config.Token = token
	}

	if err := config.writeTo(f); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Printf("Saved config to: %s\n", config.ConfigFile)
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
