package conn

import (
	"crypto/tls"
	"errors"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

type ConnectionManager interface {
	AddConnection(endPoint string) (PeerConnection, error)
	GetConnection(endPoint string) (PeerConnection, error)
}

// ConnectionManager manages connections between other peers
// in the control plane.
type JwtConnectionManager struct {
	// clientConnections maps an endpoint to a connection
	clientConnections map[string]PeerConnection
	serverConfig      *tls.Config
	clientConfig      *tls.Config
}

type NewJwtConnectionManagerParams struct {
	CertificatePath      string
	PrivateKey           string
	SkipCertVerification bool
}

func NewJwtConnectionManager(params *NewJwtConnectionManagerParams) (ConnectionManager, error) {
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

	serverConfig := &tls.Config{
		ClientAuth:   serverAuth,
		Certificates: []tls.Certificate{cert},
	}

	clientConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: params.SkipCertVerification,
	}

	connections := make(map[string]PeerConnection)
	connMgr := JwtConnectionManager{connections, serverConfig, clientConfig}
	return &connMgr, nil
}

func (m *JwtConnectionManager) GetConnection(endpoint string) (PeerConnection, error) {
	conn, exists := m.clientConnections[endpoint]

	if !exists {
		return nil, errors.New("endpoint: " + endpoint + " does not exist")
	}

	return conn, nil
}

// AddToken: Adds a connection to the list of connections to manage
func (m *JwtConnectionManager) AddConnection(endPoint string) (PeerConnection, error) {
	_, exists := m.clientConnections[endPoint]

	if exists {
		return nil, errors.New("token already exists in the connections")
	}

	connections, err := NewWgCtrlConnection(m.clientConfig, endPoint)

	if err != nil {
		return nil, err
	}

	m.clientConnections[endPoint] = connections
	return connections, nil
}