// conn manages gRPC connections between peers.
// Includes timers.
package conn

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// PeerConnection interfacing for a secure connection between
// two peers.
type PeerConnection interface {
	Connect() error
	Close() error
	Authenticate(meshId string) error
	GetClient() (*grpc.ClientConn, error)
	CreateAuthContext(meshId string) (context.Context, error)
}

type WgCtrlConnection struct {
	clientConfig *tls.Config
	conn         *grpc.ClientConn
	endpoint     string
	// tokens maps a meshID to the corresponding token
	tokens map[string]string
}

func NewWgCtrlConnection(clientConfig *tls.Config, server string) (*WgCtrlConnection, error) {
	var conn WgCtrlConnection
	conn.tokens = make(map[string]string)
	conn.clientConfig = clientConfig
	conn.endpoint = server
	return &conn, nil
}

func (c *WgCtrlConnection) Authenticate(meshId string) error {
	conn, err := grpc.Dial(c.endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(c.clientConfig)))

	defer conn.Close()

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	client := rpc.NewAuthenticationClient(conn)
	defer cancel()

	authRequest := rpc.JoinAuthMeshRequest{
		MeshId: meshId,
		Alias:  lib.GetOutboundIP().String(),
	}

	reply, err := client.JoinMesh(ctx, &authRequest)

	if err != nil {
		return err
	}

	c.tokens[meshId] = *reply.Token
	return nil
}

// ConnectWithToken: Connects to a new gRPC peer given the address of the other server.
func (c *WgCtrlConnection) Connect() error {
	conn, err := grpc.Dial(c.endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(c.clientConfig)),
	)

	if err != nil {
		logging.ErrorLog.Printf("Could not connect: %s\n", err.Error())
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
		err = errors.New("The client's config does not exist")
	}

	return c.conn, err
}

// TODO: Implement a mechanism to attach a security token
func (c *WgCtrlConnection) CreateAuthContext(meshId string) (context.Context, error) {
	token, ok := c.tokens[meshId]

	if !ok {
		return nil, errors.New("MeshID: " + meshId + " does not exist")
	}

	ctx := context.Background()
	return metadata.AppendToOutgoingContext(ctx, "authorization", token), nil
}
