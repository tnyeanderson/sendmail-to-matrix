package pkg

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/rs/zerolog"
	"go.mau.fi/gomuks/pkg/hicli"
	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	_ "github.com/glebarez/go-sqlite"
)

// Client provides an interface for an underlying Matrix client, which could be
// encrypted or unencrypted.
type Client interface {
	// Close will gracefully close the connection and client. It does not log
	// out. Trying to use the Client after calling this may not work. You must
	// run this when you are done with your Client (or before your program
	// exits).
	Close() error

	// SendAttachment will upload the attachment and send it to a room.
	SendAttachment(ctx context.Context, room string, attachment Attachment) error

	// SendText will send a text-based message to a room.
	SendText(ctx context.Context, room string, text []byte) error
}

// NewUnencryptedClient returns a Client that provides unencrypted messaging.
func NewUnencryptedClient(server, userID, token string) (Client, error) {
	mautrixClient, err := mautrix.NewClient(server, id.UserID(userID), token)
	if err != nil {
		return nil, err
	}
	return &unencryptedClient{mautrixClient}, nil
}

// NewEncryptedClient returns a Client that provides encrypted messaging. This
// function expects that the state database and configuration is already
// present on disk, so make sure that [SetupNewEncryptedClientUsingRecoveryKey]
// has succeeded previously before trying to use this function to utilize that
// Client.
func NewEncryptedClient(userID, databasePath, databasePassword string) (Client, error) {
	return newEncryptedClient(userID, databasePath, databasePassword)
}

// unencryptedClient is a [Client] which uses an access token and a
// [mautrix.Client] to provide unencrypted messaging.
type unencryptedClient struct {
	mautrixClient *mautrix.Client
}

func (c *unencryptedClient) Close() error {
	// Only needed for encrypted clients
	return nil
}

func (c *unencryptedClient) SendText(ctx context.Context, room string, text []byte) error {
	if _, err := c.mautrixClient.SendText(ctx, id.RoomID(room), string(text)); err != nil {
		return err
	}
	return nil
}

func (c *unencryptedClient) SendAttachment(ctx context.Context, room string, attachment Attachment) error {
	return sendAttachment(ctx, c.mautrixClient, room, attachment)
}

// encryptedClient is a [Client] that uses a [hicli.HiClient] to provide
// encrypted messaging.
type encryptedClient struct {
	hiClient *hicli.HiClient

	// synced is used to communicate SyncComplete events from hicli
	synced chan bool
}

func newEncryptedClient(userID, databasePath, databasePassword string) (*encryptedClient, error) {
	c := &encryptedClient{
		synced: make(chan bool),
	}
	rawDB, err := dbutil.NewWithDialect(databasePath, "sqlite")
	if err != nil {
		return nil, err
	}

	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	eventCallback := func(a any) {
		switch a.(type) {
		case *hicli.SyncComplete:
			c.synced <- true
		}
	}

	c.hiClient = hicli.New(rawDB, nil, logger, []byte(databasePassword), eventCallback)

	if err := c.hiClient.Start(context.Background(), id.UserID(userID), nil); err != nil {
		return nil, err
	}

	runtime.AddCleanup(c, func(_ bool) { c.Close() }, true)

	if !c.hiClient.IsLoggedIn() {
		return nil, fmt.Errorf("encrypted client is not logged in")
	}

	return c, nil
}

func (c *encryptedClient) Close() error {
	if c.hiClient != nil {
		c.hiClient.Stop()
	}
	return nil
}

func (c *encryptedClient) SendText(ctx context.Context, room string, message []byte) error {
	body := &event.MessageEventContent{
		Body:    string(message),
		MsgType: event.MsgText,
	}

	if _, err := c.hiClient.Send(ctx, id.RoomID(room), event.EventMessage, body, false, true); err != nil {
		return err
	}

	c.waitForSync()
	return nil
}

func (c *encryptedClient) SendAttachment(ctx context.Context, room string, attachment Attachment) error {
	return sendAttachment(ctx, c.hiClient.Client, room, attachment)
}

func (c *encryptedClient) waitForSync() {
	// TODO: This seems to sometimes cause infinite hangs after all parts have
	// been sent. Perhaps need a more reliable way to trigger a sync, wait for
	// it, then return.
	<-c.synced
}

// SetupNewEncryptedClientUsingRecoveryKey performs login and device
// verification, then return the Client for use. The state database is saved to
// disk, and can be loaded later using [NewEncryptedClient].
func SetupNewEncryptedClientUsingRecoveryKey(
	ctx context.Context,
	server, user, password, recoveryCode,
	databasePath, databasePassword, deviceName string,
) (Client, error) {
	c, err := newEncryptedClient(user, databasePath, databasePassword)
	if err != nil {
		return nil, err
	}
	hicli.InitialDeviceDisplayName = deviceName
	if err := c.hiClient.LoginAndVerify(ctx, server, user, password, recoveryCode); err != nil {
		return nil, err
	}
	c.waitForSync()
	return c, nil
}

func sendAttachment(ctx context.Context, mautrixClient *mautrix.Client, room string, attachment Attachment) error {
	uploadRes, err := mautrixClient.UploadBytes(ctx, attachment.Content, attachment.ContentType)
	if err != nil {
		return err
	}
	fileName := attachment.FileName
	if fileName == "" {
		fileName = "attachment"
	}
	_, err = mautrixClient.SendMessageEvent(ctx, id.RoomID(room), event.EventMessage, &event.MessageEventContent{
		Body:    fileName,
		MsgType: event.MsgImage,
		URL:     uploadRes.ContentURI.CUString(),
		Info:    &event.FileInfo{MimeType: attachment.ContentType},
	})
	return err
}
