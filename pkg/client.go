package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/hicli"
	"maunium.net/go/mautrix/id"
)

// UnencryptedClient sends unencrypted messages to a matrix room.
type UnencryptedClient struct {
	Server string
	Token  string
}

func NewUnencryptedClient(server, token string) (*UnencryptedClient, error) {
	return &UnencryptedClient{
		Server: server,
		Token:  token,
	}, nil
}

// SendMessage sends an unencrypted message to a matrix room.
func (c *UnencryptedClient) SendMessage(ctx context.Context, room string, message []byte) error {
	urlFmt := "%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s"
	transactionID, err := generateTransactionID()
	if err != nil {
		return err
	}
	url := fmt.Sprintf(urlFmt, c.Server, room, transactionID)
	body := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(body)
	enc.SetEscapeHTML(false)
	err = enc.Encode(matrixRequestBody{
		Body:    string(message),
		Msgtype: "m.text",
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Set("access_token", c.Token)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	_, err = client.Do(req)
	return err
}

// EncryptedClient sends encrypted messages to a matrix room.
type EncryptedClient struct {
	hicli  *hicli.HiClient
	synced *bool
}

func NewEncryptedClient(ctx context.Context, dbPath string, picklePass string, logger zerolog.Logger) (*EncryptedClient, error) {
	c := &EncryptedClient{}
	rawDB, err := dbutil.NewWithDialect(dbPath, "sqlite3-fk-wal")
	if err != nil {
		return nil, err
	}

	s := false
	c.synced = &s
	c.hicli = hicli.New(rawDB, nil, logger, []byte(picklePass), func(a any) {
		switch a.(type) {
		case *hicli.SyncComplete:
			*c.synced = true
		}
	})

	return c, nil
}

// SendMessage sends an encrypted message to a matrix room.
func (c *EncryptedClient) SendMessage(ctx context.Context, room string, message []byte) error {
	userID, err := c.hicli.DB.Account.GetFirstUserID(ctx)
	if err != nil {
		return err
	}

	if err := c.hicli.Start(ctx, userID, nil); err != nil {
		return err
	}

	if !c.hicli.IsLoggedIn() {
		return fmt.Errorf("Not logged in")
	}

	body := &event.MessageEventContent{
		Body:    string(message),
		MsgType: event.MsgText,
	}

	if _, err := c.hicli.Send(ctx, id.RoomID(room), event.EventMessage, body); err != nil {
		return err
	}

	c.waitForSync(ctx)
	c.hicli.Stop()
	return nil
}

// LoginAndVerify authenticates a user and performs device verification,
// allowing encrypted messages. It should only need to be run once to during
// initial setup.
func (c *EncryptedClient) LoginAndVerify(ctx context.Context, server, user, password, recoveryCode, deviceName string) error {
	hicli.InitialDeviceDisplayName = deviceName
	if err := c.hicli.Start(ctx, id.UserID(user), nil); err != nil {
		return err
	}
	if err := c.hicli.LoginAndVerify(ctx, server, user, password, recoveryCode); err != nil {
		return err
	}
	c.waitForSync(ctx)
	c.hicli.Stop()
	return nil
}

func (c *EncryptedClient) waitForSync(ctx context.Context) {
	*c.synced = false
	for {
		if *c.synced {
			break
		}
	}
}
