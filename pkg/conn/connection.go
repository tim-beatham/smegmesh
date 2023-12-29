// conn manages gRPC connections between peers.
// Includes timers.
package conn

import (
	"crypto/tls"
	"errors"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// PeerConnection represents a client-side connection between two
// peers.
type PeerConnection interface {
	Close() error
	GetClient() (*grpc.ClientConn, error)
}

type PeerConnectionFactory = func(clientConfig *tls.Config, server string) (PeerConnection, error)

// WgCtrlConnection implements PeerConnection.
type WgCtrlConnection struct {
	clientConfig *tls.Config
	conn         *grpc.ClientConn
	endpoint     string
}

// NewWgCtrlConnection creates a new instance of a WireGuard control connection
func NewWgCtrlConnection(clientConfig *tls.Config, server string) (PeerConnection, error) {
	var conn WgCtrlConnection
	conn.clientConfig = clientConfig
	conn.endpoint = server

	if err := conn.CreateGrpcConnection(); err != nil {
		return nil, err
	}

	return &conn, nil
}

// ConnectWithToken: Connects to a new gRPC peer given the address of the other server.
func (c *WgCtrlConnection) CreateGrpcConnection() error {
	retryPolicy := `{
		"methodConfig": [{
		  "name": [
			{"service": "syncservice.SyncService"},
			{"service": "ctrlserver.MeshCtrlServer"}
		  ],
		  "waitForReady": true,
		  "retryPolicy": {
			  "MaxAttempts": 2,
			  "InitialBackoff": ".01s",
			  "MaxBackoff": ".01s",
			  "BackoffMultiplier": 1.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`

	conn, err := grpc.Dial(c.endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(c.clientConfig)),
		grpc.WithDefaultServiceConfig(retryPolicy))

	if err != nil {
		logging.Log.WriteErrorf("Could not connect: %s\n", err.Error())
		return err
	}

	c.conn = conn
	return nil
}

// Close: Closes the client connections
func (c *WgCtrlConnection) Close() error {
	return c.conn.Close()
}

// GetClient: Gets the client connection
func (c *WgCtrlConnection) GetClient() (*grpc.ClientConn, error) {
	var err error = nil

	if c.conn == nil {
		err = errors.New("the client's config does not exist")
	}

	return c.conn, err
}
