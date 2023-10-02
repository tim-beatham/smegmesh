package conn

import (
	"crypto/tls"
	"errors"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

// ConnectionManager manages connections between other peers
// in the control plane.
type ConnectionManager struct {
	// clientConnections maps an endpoint to a connection
	clientConnections map[string]PeerConnection
	serverConfig      *tls.Config
	clientConfig      *tls.Config
}

type NewConnectionManagerParams struct {
	CertificatePath      string
	PrivateKey           string
	SkipCertVerification bool
}

func NewConnectionManager(params *NewConnectionManagerParams) (*ConnectionManager, error) {
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
	connMgr := ConnectionManager{connections, serverConfig, clientConfig}
	return &connMgr, nil
}

func (m *ConnectionManager) GetConnection(endpoint string) (PeerConnection, error) {
	conn, exists := m.clientConnections[endpoint]

	if !exists {
		return nil, errors.New("endpoint: " + endpoint + " does not exist")
	}

	return conn, nil
}

type AddConnectionParams struct {
	TokenId string
}

// AddToken: Adds a connection to the list of connections to manage
func (m *ConnectionManager) AddConnection(endPoint string) (PeerConnection, error) {
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
