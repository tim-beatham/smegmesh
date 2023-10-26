package mesh

import (
	"errors"
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshManager struct {
	Meshes       map[string]MeshProvider
	RouteManager RouteManager
	Client       *wgctrl.Client
	// HostParameters contains information that uniquely locates
	// the node in the mesh network.
	HostParameters      *HostParameters
	conf                *conf.WgMeshConfiguration
	meshProviderFactory MeshProviderFactory
	configApplyer       MeshConfigApplyer
	interfaceEnabler    MeshInterface
}

// CreateMesh: Creates a new mesh, stores it and returns the mesh id
func (m *MeshManager) CreateMesh(devName string, port int) (string, error) {
	key, err := wgtypes.GenerateKey()

	if err != nil {
		return "", err
	}

	nodeManager, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName: devName,
		Port:    port,
		Conf:    m.conf,
		Client:  m.Client,
		MeshId:  key.String(),
	})

	if err != nil {
		return "", err
	}

	m.Meshes[key.String()] = nodeManager

	return key.String(), err
}

// AddMesh: Add the mesh to the list of meshes
func (m *MeshManager) AddMesh(meshId string, devName string, port int, meshBytes []byte) error {
	meshProvider, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName: devName,
		Port:    port,
		Conf:    m.conf,
		Client:  m.Client,
		MeshId:  meshId,
	})

	if err != nil {
		return err
	}

	err = meshProvider.Load(meshBytes)

	if err != nil {
		return err
	}

	m.Meshes[meshId] = meshProvider
	return nil
}

// AddMeshNode: Add a mesh node
func (m *MeshManager) AddMeshNode(meshId string, node MeshNode) {
	m.Meshes[meshId].AddNode(node)
}

// HasChanges returns true if the mesh has changes
func (m *MeshManager) HasChanges(meshId string) bool {
	return m.Meshes[meshId].HasChanges()
}

// GetMesh returns the mesh with the given meshid
func (m *MeshManager) GetMesh(meshId string) MeshProvider {
	theMesh, _ := m.Meshes[meshId]
	return theMesh
}

// EnableInterface: Enables the given WireGuard interface.
func (s *MeshManager) EnableInterface(meshId string) error {
	err := s.configApplyer.ApplyConfig()

	if err != nil {
		return err
	}

	return s.interfaceEnabler.EnableInterface(meshId)
}

// GetPublicKey: Gets the public key of the WireGuard mesh
func (s *MeshManager) GetPublicKey(meshId string) (*wgtypes.Key, error) {
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

func (s *MeshManager) GetSelf(meshId string) (MeshNode, error) {
	meshInstance, ok := s.Meshes[meshId]

	if !ok {
		return nil, errors.New(fmt.Sprintf("mesh %s does not exist", meshId))
	}

	snapshot, err := meshInstance.GetMesh()

	if err != nil {
		return nil, err
	}

	node, ok := snapshot.GetNodes()[s.HostParameters.HostEndpoint]

	if !ok {
		return nil, errors.New("the node doesn't exist in the mesh")
	}

	return node, nil
}

// UpdateTimeStamp updates the timestamp of this node in all meshes
func (s *MeshManager) UpdateTimeStamp() error {
	for _, mesh := range s.Meshes {
		err := mesh.UpdateTimeStamp(s.HostParameters.HostEndpoint)

		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a new instance of a mesh manager with the given parameters
func NewMeshManager(conf conf.WgMeshConfiguration, client *wgctrl.Client, meshProvider MeshProviderFactory) *MeshManager {
	hostParams := HostParameters{}

	switch conf.PublicEndpoint {
	case "":
		hostParams.HostEndpoint = fmt.Sprintf("%s:%s", lib.GetOutboundIP().String(), conf.GrpcPort)
	default:
		hostParams.HostEndpoint = fmt.Sprintf("%s:%s", conf.PublicEndpoint, conf.GrpcPort)
	}

	logging.Log.WriteInfof("Endpoint %s", hostParams.HostEndpoint)

	m := &MeshManager{
		Meshes:              make(map[string]MeshProvider),
		HostParameters:      &hostParams,
		meshProviderFactory: meshProvider,
		Client:              client,
		conf:                &conf,
	}
	m.configApplyer = NewWgMeshConfigApplyer(m)
	m.RouteManager = NewRouteManager(m)
	m.interfaceEnabler = NewWgMeshInterface(m)
	return m
}
