package conn

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"sync"

	logging "github.com/tim-beatham/smegmesh/pkg/log"
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
	// HasConnections returns true if a peer has already registered at the given
	// endpoint or false otherwise.
	HasConnection(endPoint string) bool
	// Removes a connection if it exists
	RemoveConnection(endPoint string) error
	// Goes through all the connections and closes eachone
	Close() error
}

// ConnectionManager manages connections between other peers
// in the control plane.
type ConnectionManagerImpl struct {
	// clientConnections maps an endpoint to a connection
	conLoc            sync.RWMutex
	clientConnections map[string]PeerConnection
	clientConfig      *tls.Config
	connFactory       PeerConnectionFactory
}

// Create a new instance of a connection manager.
type NewConnectionManagerParams struct {
	// The path to the certificate
	CertificatePath string
	// The private key of the node
	PrivateKey string
	// Whether or not to skip certificate verification
	SkipCertVerification bool
	CaCert               string
	ConnFactory          PeerConnectionFactory
}

// NewConnectionManager: Creates a new instance of a ConnectionManager or an error
// if something went wrong.
func NewConnectionManager(params *NewConnectionManagerParams) (ConnectionManager, error) {
	cert, err := tls.LoadX509KeyPair(params.CertificatePath, params.PrivateKey)

	if err != nil {
		logging.Log.WriteErrorf("Failed to load key pair: %s\n", err.Error())
		logging.Log.WriteErrorf("Certificate Path: %s\n", params.CertificatePath)
		logging.Log.WriteErrorf("Private Key Path: %s\n", params.PrivateKey)
		return nil, err
	}

	certPool := x509.NewCertPool()

	if params.CaCert == "" {
		return nil, errors.New("CA Cert is not specified")
	}

	caCert, err := os.ReadFile(params.CaCert)

	if err != nil {
		return nil, err
	}

	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("could not parse PEM")
	}

	clientConfig := &tls.Config{
		InsecureSkipVerify: params.SkipCertVerification,
		Certificates:       []tls.Certificate{cert},
		RootCAs:            certPool,
	}

	connections := make(map[string]PeerConnection)
	connMgr := ConnectionManagerImpl{
		sync.RWMutex{},
		connections,
		clientConfig,
		params.ConnFactory,
	}

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

	connections, err := m.connFactory(m.clientConfig, endPoint)

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

// RemoveConnection removes the given connection if it exists
func (m *ConnectionManagerImpl) RemoveConnection(endPoint string) error {
	m.conLoc.Lock()
	connection, ok := m.clientConnections[endPoint]

	var err error

	if ok {
		err = connection.Close()
		delete(m.clientConnections, endPoint)
	}

	m.conLoc.Unlock()
	return err
}
func (m *ConnectionManagerImpl) Close() error {
	for _, conn := range m.clientConnections {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
