package conn

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/auth"
	"github.com/tim-beatham/wgmesh/pkg/conf"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ConnectionServer manages the gRPC server
type ConnectionServer struct {
	severConfig  *tls.Config
	JwtManager   *auth.JwtManager
	server       *grpc.Server
	authProvider rpc.AuthenticationServer
	ctrlProvider rpc.MeshCtrlServerServer
	Conf         *conf.WgMeshConfiguration
}

type NewConnectionServerParams struct {
	Conf         *conf.WgMeshConfiguration
	AuthProvider rpc.AuthenticationServer
	CtrlProvider rpc.MeshCtrlServerServer
}

// NewConnectionServer: create a new gRPC connection server instance
func NewConnectionServer(params *NewConnectionServerParams) (*ConnectionServer, error) {
	cert, err := tls.LoadX509KeyPair(params.Conf.CertificatePath, params.Conf.PrivateKeyPath)

	if err != nil {
		logging.ErrorLog.Printf("Failed to load key pair: %s\n", err.Error())
		logging.ErrorLog.Printf("Certificate Path: %s\n", params.Conf.CertificatePath)
		logging.ErrorLog.Printf("Private Key Path: %s\n", params.Conf.PrivateKeyPath)
		return nil, err
	}

	serverAuth := tls.RequireAndVerifyClientCert

	if params.Conf.SkipCertVerification {
		serverAuth = tls.RequireAnyClientCert
	}

	serverConfig := &tls.Config{
		ClientAuth:   serverAuth,
		Certificates: []tls.Certificate{cert},
	}

	jwtManager := auth.NewJwtManager("tim123", 24*time.Hour)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(jwtManager.GetAuthInterceptor()),
		grpc.Creds(credentials.NewTLS(serverConfig)),
	)

	authProvider := params.AuthProvider
	ctrlProvider := params.CtrlProvider

	connServer := ConnectionServer{
		serverConfig,
		jwtManager,
		server,
		authProvider,
		ctrlProvider,
		params.Conf,
	}

	return &connServer, nil
}

func (s *ConnectionServer) Listen() error {
	rpc.RegisterMeshCtrlServerServer(s.server, s.ctrlProvider)
	rpc.RegisterAuthenticationServer(s.server, s.authProvider)

	lis, err := net.Listen("tcp", s.Conf.GrpcPort)

	if err != nil {
		logging.ErrorLog.Println(err.Error())
		return err
	}

	if err := s.server.Serve(lis); err != nil {
		logging.ErrorLog.Println(err.Error())
		return err
	}

	return nil
}
