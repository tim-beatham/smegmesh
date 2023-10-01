// conn manages gRPC connections between peers.
// Includes timers.
package conn

import (
	"crypto/tls"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// PeerConnection interfacing for a secure connection between
// two peers.
type PeerConnection interface {
	Connect() error
}

type WgCtrlConnection struct {
	serverConfig *tls.Config
	clientConfig *tls.Config
	conn         *grpc.ClientConn
}

type NewConnectionsParams struct {
	CertificatePath      string
	PrivateKey           string
	SkipCertVerification bool
}

func NewConnection(params *NewConnectionsParams) (*WgCtrlConnection, error) {
	cert, err := tls.LoadX509KeyPair(params.CertificatePath, params.PrivateKey)

	if err != nil {
		logging.ErrorLog.Printf("Failed to load key pair: %s\n", err.Error())
		logging.ErrorLog.Printf("Certificate Path: %s\n", params.CertificatePath)
		logging.ErrorLog.Printf("Private Key Path: %s\n", params.PrivateKey)
		return nil, err
	}

	serverAuth := tls.RequireAndVerifyClientCert

	if params.SkipCertVerification {
		serverAuth = tls.RequireAnyClientCert
	}

	tlsConfig := &tls.Config{
		ClientAuth:   serverAuth,
		Certificates: []tls.Certificate{cert},
	}

	clientConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: params.SkipCertVerification,
	}

	wgConnection := WgCtrlConnection{serverConfig: tlsConfig, clientConfig: clientConfig}

	return &wgConnection, nil
}

// Connect: Connects to a new gRPC peer given the address of the other server
func (c *WgCtrlConnection) Connect(server string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(credentials.NewTLS(c.clientConfig)))

	if err != nil {
		logging.ErrorLog.Printf("Could not connect: %s\n", err.Error())
		return nil, err
	}

	return conn, nil
}

// Listen: listens to incoming messages
func (c *WgCtrlConnection) Listen(i grpc.UnaryServerInterceptor) *grpc.Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(i),
		grpc.Creds(credentials.NewTLS(c.serverConfig)),
	)
	return server
}
