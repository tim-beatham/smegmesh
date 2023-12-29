package conn

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ConnectionServer manages gRPC server peer connections
type ConnectionServer struct {
	// server an instance of the grpc server
	server *grpc.Server
	// the ctrl service to manage node
	ctrlProvider rpc.MeshCtrlServerServer
	// the sync service to synchronise nodes
	syncProvider rpc.SyncServiceServer
	Conf         *conf.DaemonConfiguration
	listener     net.Listener
}

// NewConnectionServerParams contains params for creating a new connection server
type NewConnectionServerParams struct {
	Conf         *conf.DaemonConfiguration
	CtrlProvider rpc.MeshCtrlServerServer
	SyncProvider rpc.SyncServiceServer
}

// NewConnectionServer: create a new gRPC connection server instance
func NewConnectionServer(params *NewConnectionServerParams) (*ConnectionServer, error) {
	cert, err := tls.LoadX509KeyPair(params.Conf.CertificatePath, params.Conf.PrivateKeyPath)

	if err != nil {
		logging.Log.WriteErrorf("Failed to load key pair: %s\n", err.Error())
		return nil, err
	}

	serverAuth := tls.RequireAndVerifyClientCert

	if params.Conf.SkipCertVerification {
		serverAuth = tls.RequireAnyClientCert
	}

	certPool := x509.NewCertPool()

	if params.Conf.CaCertificatePath == "" {
		return nil, errors.New("CA Cert is not specified")
	}

	caCert, err := os.ReadFile(params.Conf.CaCertificatePath)

	if err != nil {
		return nil, err
	}

	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("could not parse PEM")
	}

	serverConfig := &tls.Config{
		ClientAuth:   serverAuth,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
	}

	server := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(serverConfig)),
	)

	ctrlProvider := params.CtrlProvider
	syncProvider := params.SyncProvider

	connServer := ConnectionServer{
		server:       server,
		ctrlProvider: ctrlProvider,
		syncProvider: syncProvider,
		Conf:         params.Conf,
	}

	return &connServer, nil
}

// Listen for incoming requests. Returns an error if something went wrong.
func (s *ConnectionServer) Listen() error {
	rpc.RegisterMeshCtrlServerServer(s.server, s.ctrlProvider)
	rpc.RegisterSyncServiceServer(s.server, s.syncProvider)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Conf.GrpcPort))
	s.listener = lis

	logging.Log.WriteInfof("GRPC listening on %d\n", s.Conf.GrpcPort)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return err
	}

	if err := s.server.Serve(lis); err != nil {
		logging.Log.WriteErrorf(err.Error())
		return err
	}

	return nil
}

// Close closes the connection server. Returns an error
// if something went wrong whilst attempting to close the connection
func (c *ConnectionServer) Close() error {
	var err error = nil
	c.server.Stop()

	if c.listener != nil {
		err = c.listener.Close()
	}

	return err
}
