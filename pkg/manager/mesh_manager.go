package manager

import (
	"errors"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
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
func (m *MeshManger) CreateMesh(devName string) (string, error) {
	key, err := wgtypes.GenerateKey()

	if err != nil {
		return "", err
	}

	nodeManager := crdt.NewCrdtNodeManager(key.String(), devName, m.Client)
	m.Meshes[key.String()] = nodeManager
	return key.String(), nil
}

// UpdateMesh: merge the changes and save it to the device
func (m *MeshManger) UpdateMesh(meshId string, changes []byte) error {
	mesh, ok := m.Meshes[meshId]

	if !ok {
		return errors.New("mesh does not exist")
	}

	err := mesh.LoadChanges(changes)

	if err != nil {
		return err
	}

	return nil
}

// AddMesh: Add the mesh to the list of meshes
func (m *MeshManger) AddMesh(meshId string, devName string, meshBytes []byte) error {
	mesh := crdt.NewCrdtNodeManager(meshId, devName, m.Client)
	err := mesh.Load(meshBytes)

	if err != nil {
		return err
	}

	m.Meshes[meshId] = mesh
	return nil
}

// AddMeshNode: Add a mesh node
func (m *MeshManger) AddMeshNode(meshId string, node crdt.MeshNodeCrdt) {
	m.Meshes[meshId].AddNode(node)
}

func (m *MeshManger) GetMesh(meshId string) *crdt.CrdtNodeManager {
	theMesh, _ := m.Meshes[meshId]
	return theMesh
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

	dev, err := s.Client.Device(mesh.IfName)

	if err != nil {
		return err
	}

	node, contains := crdt.Nodes[dev.PublicKey.String()]

	if !contains {
		return errors.New("Node does not exist in the mesh")
	}

	return wg.EnableInterface(mesh.IfName, node.WgHost)
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

func NewMeshManager(client wgctrl.Client) *MeshManger {
	return &MeshManger{Meshes: make(map[string]*crdt.CrdtNodeManager), Client: &client}
}
