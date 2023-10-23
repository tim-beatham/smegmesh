package conn

import (
	"crypto/tls"
	"sync"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

// ConnectionManager defines an interface for maintaining peer connections
type ConnectionManager interface {
	// AddConnection adds an instance of a connection at the given endpoint
	// or error if something went wrong
	AddConnection(endPoint string) (PeerConnection, error)
	// GetConnection returns an instance of a connection at the given endpoint.
	// If the endpoint does not exist then add the connection. Returns an error
	// if something went wrong
	GetConnection(endPoint string) (PeerConnection, error)
	// HasConnections returns true if a client has already registered at the givne
	// endpoint or false otherwise.
	HasConnection(endPoint string) bool
	// Goes through all the connections and closes eachone
	Close() error
}

// ConnectionManager manages connections between other peers
// in the control plane.
type ConnectionManagerImpl struct {
	// clientConnections maps an endpoint to a connection
	conLoc            sync.RWMutex
	clientConnections map[string]PeerConnection
	serverConfig      *tls.Config
	clientConfig      *tls.Config
}

// Create a new instance of a connection manager.
type NewConnectionManageParams struct {
	// The path to the certificate
	CertificatePath string
	// The private key of the node
	PrivateKey string
	// Whether or not to skip certificate verification
	SkipCertVerification bool
}

// NewConnectionManager: Creates a new instance of a ConnectionManager or an error
// if something went wrong.
func NewConnectionManager(params *NewConnectionManageParams) (ConnectionManager, error) {
	cert, err := tls.LoadX509KeyPair(params.CertificatePath, params.PrivateKey)

	if err != nil {
		logging.Log.WriteErrorf("Failed to load key pair: %s\n", err.Error())
		logging.Log.WriteErrorf("Certificate Path: %s\n", params.CertificatePath)
		logging.Log.WriteErrorf("Private Key Path: %s\n", params.PrivateKey)
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
	connMgr := ConnectionManagerImpl{sync.RWMutex{}, connections, serverConfig, clientConfig}
	return &connMgr, nil
}

// GetConnection: Returns the given connection if it exists. If it does not exist then add
// the connection. Returns an error if something went wrong
func (m *ConnectionManagerImpl) GetConnection(endpoint string) (PeerConnection, error) {
	m.conLoc.Lock()
	conn, exists := m.clientConnections[endpoint]
	m.conLoc.Unlock()

	if !exists {
		return m.AddConnection(endpoint)
	}

	return conn, nil
}

// AddConnection: Adds a connection to the list of connections to manage.
func (m *ConnectionManagerImpl) AddConnection(endPoint string) (PeerConnection, error) {
	m.conLoc.Lock()
	conn, exists := m.clientConnections[endPoint]
	m.conLoc.Unlock()

	if exists {
		return conn, nil
	}

	connections, err := NewWgCtrlConnection(m.clientConfig, endPoint)

	if err != nil {
		return nil, err
	}

	m.conLoc.Lock()
	m.clientConnections[endPoint] = connections
	m.conLoc.Unlock()
	return connections, nil
}

// HasConnection Returns TRUE if the given endpoint exists
func (m *ConnectionManagerImpl) HasConnection(endPoint string) bool {
	_, exists := m.clientConnections[endPoint]
	return exists
}

func (m *ConnectionManagerImpl) Close() error {
	for _, conn := range m.clientConnections {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}