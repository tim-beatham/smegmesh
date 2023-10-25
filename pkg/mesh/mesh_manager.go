package mesh

import (
	"errors"
	"fmt"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshManger struct {
	Meshes       map[string]*crdt.CrdtNodeManager
	RouteManager RouteManager
	Client       *wgctrl.Client
	HostEndpoint string
	conf         *conf.WgMeshConfiguration
}

func (m *MeshManger) MeshExists(meshId string) bool {
	_, inMesh := m.Meshes[meshId]
	return inMesh
}

// CreateMesh: Creates a new mesh, stores it and returns the mesh id
func (m *MeshManger) CreateMesh(devName string, port int) (string, error) {
	key, err := wgtypes.GenerateKey()

	if err != nil {
		return "", err
	}

	nodeManager, err := crdt.NewCrdtNodeManager(key.String(), m.HostEndpoint, devName, port, *m.conf, m.Client)

	if err != nil {
		return "", err
	}

	m.Meshes[key.String()] = nodeManager

	return key.String(), err
}

// AddMesh: Add the mesh to the list of meshes
func (m *MeshManger) AddMesh(meshId string, devName string, port int, meshBytes []byte) error {
	mesh, err := crdt.NewCrdtNodeManager(meshId, m.HostEndpoint, devName, port, *m.conf, m.Client)

	if err != nil {
		return err
	}

	err = mesh.Load(meshBytes)

	if err != nil {
		return err
	}

	m.Meshes[meshId] = mesh
	return nil
}

// AddMeshNode: Add a mesh node
func (m *MeshManger) AddMeshNode(meshId string, node crdt.MeshNodeCrdt) {
	m.Meshes[meshId].AddNode(node)

	if m.conf.AdvertiseRoutes {
		m.RouteManager.UpdateRoutes()
	}
}

func (m *MeshManger) HasChanges(meshId string) bool {
	return m.Meshes[meshId].HasChanges()
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

	node, contains := crdt.Nodes[s.HostEndpoint]

	if !contains {
		return errors.New("Node does not exist in the mesh")
	}

	err = mesh.ApplyWg()

	if err != nil {
		return err
	}

	err = wg.EnableInterface(mesh.IfName, node.WgHost)

	if s.conf.AdvertiseRoutes {
		s.RouteManager.ApplyWg(mesh)
	}

	return nil
}

// GetPublicKey: Gets the public key of the WireGuard mesh
func (s *MeshManger) GetPublicKey(meshId string) (*wgtypes.Key, error) {
	mesh, ok := s.Meshes[meshId]

	if !ok {
		return nil, errors.New("mesh does not exist")
	}

	dev, err := mesh.GetDevice()

	if err != nil {
		return nil, err
	}

	return &dev.PublicKey, nil
}

// UpdateTimeStamp updates the timestamp of this node in all meshes
func (s *MeshManger) UpdateTimeStamp() error {
	for _, mesh := range s.Meshes {
		err := mesh.UpdateTimeStamp()

		if err != nil {
			return err
		}
	}

	return nil
}

func NewMeshManager(conf conf.WgMeshConfiguration, client *wgctrl.Client) *MeshManger {
	ip := lib.GetOutboundIP()
	m := &MeshManger{
		Meshes:       make(map[string]*crdt.CrdtNodeManager),
		HostEndpoint: fmt.Sprintf("%s:%s", ip.String(), conf.GrpcPort),
		Client:       client,
		conf:         &conf,
	}

	m.RouteManager = NewRouteManager(m)
	return m
}
