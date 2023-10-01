package auth

import (
	"errors"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

type TokenMesh struct {
	Tokens map[string]string
}

type TokenManager struct {
	Meshes map[string]*TokenMesh
}

func (m *TokenManager) AddToken(meshId, endpoint, token string) error {
	mesh, ok := m.Meshes[endpoint]

	if !ok {
		mesh = new(TokenMesh)
		mesh.Tokens = make(map[string]string)
		m.Meshes[endpoint] = mesh
	}

	mesh.Tokens[meshId] = token
	return nil
}

func (m *TokenManager) GetToken(meshId, endpoint string) (string, error) {
	mesh, ok := m.Meshes[endpoint]

	if !ok {
		logging.ErrorLog.Printf("Endpoint doesnot exist: %s\n", endpoint)
		return "", errors.New("Endpoint does not exist in the token manager")
	}

	token, ok := mesh.Tokens[meshId]

	if !ok {
		return "", errors.New("MeshId does not exist")
	}

	return token, nil
}

func NewTokenManager() *TokenManager {
	var manager *TokenManager = new(TokenManager)
	manager.Meshes = make(map[string]*TokenMesh)
	return manager
}
