package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		c, err := getConfig()
		if err != nil {
			return err
		}

		if c.EncryptionDisabled {
			return setupWithoutEncryption(c)
		}

		return setup(c)
	},
}

func ask(prompt, defaultValue string) string {
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]: ", prompt, defaultValue)
	} else {
		prompt = prompt + ": "
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	scanner.Scan()
	answer := scanner.Text()
	if answer == "" {
		return defaultValue
	}
	return answer
}

func setup(config *cliConfig) error {
	server := ask("Matrix home server", "https://matrix.org")
	user := ask("Matrix username", "")
	password := ask("Matrix password", "")
	recoveryCode := ask("Recovery code (used for device verification)", "")
	deviceName := ask("Device display name", "sendmail-to-matrix")
	picklePass := ask("Database encryption passphrase", "sendmail-to-matrix")

	if err := os.MkdirAll(config.ConfigDir, 0750); err != nil {
		return err
	}
	dbPath := filepath.Join(config.ConfigDir, "stm.db")
	ctx := context.Background()
	logger := zerolog.New(os.Stderr)
	c, err := pkg.NewEncryptedClient(ctx, dbPath, picklePass, logger)
	if err != nil {
		return err
	}
	return c.LoginAndVerify(ctx, server, user, password, recoveryCode, deviceName)
}

func setupWithoutEncryption(config *cliConfig) error {
	// Prompt user
	server := ask("Matrix home server", "https://matrix.org")
	user := ask("Matrix username", "")
	password := ask("Matrix password", "")
	room := ask("Matrix room", "")
	preface := ask("Message preface", "")

	configFile := filepath.Join(config.ConfigDir, "config.json")

	// Fetch access_token
	token, err := pkg.GetToken(server, user, password)
	if err != nil {
		log.Fatal(err)
	}

	// Generate json
	data := map[string]string{
		"server":  server,
		"token":   token,
		"room":    room,
		"preface": preface,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Print the result
	fmt.Printf("\n%s\n\n", string(b))

	// Save to file
	if err := os.MkdirAll(config.ConfigDir, 0750); err != nil {
		return err
	}
	err = ioutil.WriteFile(configFile, b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// Notify user
	fmt.Println("")
	fmt.Printf("Saved config to: %s\n", configFile)
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
