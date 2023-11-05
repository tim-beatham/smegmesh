package conn

import (
	"crypto/tls"

	"google.golang.org/grpc"
)

type ConnectionManagerStub struct {
	Endpoints map[string]PeerConnection
}

func (s *ConnectionManagerStub) AddConnection(endPoint string) (PeerConnection, error) {
	mock := &PeerConnectionMock{}
	s.Endpoints[endPoint] = mock
	return mock, nil
}

func (s *ConnectionManagerStub) GetConnection(endPoint string) (PeerConnection, error) {
	endpoint, ok := s.Endpoints[endPoint]

	if !ok {
		return s.AddConnection(endPoint)
	}

	return endpoint, nil
}

func (s *ConnectionManagerStub) HasConnection(endPoint string) bool {
	_, ok := s.Endpoints[endPoint]
	return ok
}

func (s *ConnectionManagerStub) Close() error {
	return nil
}

type PeerConnectionMock struct {
}

func (c *PeerConnectionMock) Close() error {
	return nil
}

func (c *PeerConnectionMock) GetClient() (*grpc.ClientConn, error) {
	return &grpc.ClientConn{}, nil
}

var MockFactory PeerConnectionFactory = func(clientConfig *tls.Config, server string) (PeerConnection, error) {
	return &PeerConnectionMock{}, nil
}
