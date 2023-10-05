package manager

import (
	"errors"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshManger struct {
	Meshes map[string]*crdt.CrdtNodeManager
	Client *wgctrl.Client
}

func (m *MeshManger) MeshExists(meshId string) bool {
	_, inMesh := m.Meshes[meshId]
	return inMesh
}

// CreateMesh: Creates a new mesh, stores it and returns the mesh id
func (m *MeshManger) CreateMesh(meshId, devName string) (string, error) {
	key, err := wgtypes.GenerateKey()

	if err != nil {
		return "", err
	}

	nodeManager := crdt.NewCrdtNodeManager(key.String(), devName)
	m.Meshes[key.String()] = nodeManager
	return key.String(), nil
}

// UpdateMesh: merge the changes and save it to the device
func (m *MeshManger) UpdateMesh(meshId string, changes []byte, client wgctrl.Client) error {
	mesh, ok := m.Meshes[meshId]

	if !ok {
		return errors.New("mesh does not exist")
	}

	mesh.LoadChanges(changes)

	crdt, err := mesh.GetCrdt()

	if err != nil {
		return err
	}

	wg.UpdateWgConf(m.Meshes[meshId].IfName, crdt.Nodes, client)
	return nil
}

// EnableInterface: Enables the given WireGuard interface.
func (s *MeshManger) EnableInterface(meshId string) error {
	mesh, contains := s.Meshes[meshId]

	if !contains {
		return errors.New("Mesh does not exist")
	}

	crdt, err := mesh.GetCrdt()

	if err != nil {
		return err
	}

	endPoint := lib.GetOutboundIP().String() + ":8080"
	node, contains := crdt.Nodes[endPoint]

	if !contains {
		return errors.New("Node does not exist in the mesh")
	}

	return wg.EnableInterface(mesh.IfName, node.WgEndpoint)
}

// GetPublicKey: Gets the public key of the WireGuard mesh
func (s *MeshManger) GetPublicKey(meshId string) (*wgtypes.Key, error) {
	mesh, ok := s.Meshes[meshId]

	if !ok {
		return nil, errors.New("mesh does not exist")
	}

	dev, err := s.Client.Device(mesh.IfName)

	if err != nil {
		return nil, err
	}

	return &dev.PublicKey, nil
}
