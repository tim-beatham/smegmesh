package mesh

import (
	"errors"
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshManager interface {
	CreateMesh(devName string, port int) (string, error)
	AddMesh(params *AddMeshParams) error
	HasChanges(meshid string) bool
	GetMesh(meshId string) MeshProvider
	EnableInterface(meshId string) error
	GetPublicKey(meshId string) (*wgtypes.Key, error)
	AddSelf(params *AddSelfParams) error
	LeaveMesh(meshId string) error
	GetSelf(meshId string) (MeshNode, error)
	ApplyConfig() error
	SetDescription(description string) error
	UpdateTimeStamp() error
	GetClient() *wgctrl.Client
	GetMeshes() map[string]MeshProvider
}

type MeshManagerImpl struct {
	Meshes       map[string]MeshProvider
	RouteManager RouteManager
	Client       *wgctrl.Client
	// HostParameters contains information that uniquely locates
	// the node in the mesh network.
	HostParameters       *HostParameters
	conf                 *conf.WgMeshConfiguration
	meshProviderFactory  MeshProviderFactory
	nodeFactory          MeshNodeFactory
	configApplyer        MeshConfigApplyer
	idGenerator          lib.IdGenerator
	ipAllocator          ip.IPAllocator
	interfaceManipulator wg.WgInterfaceManipulator
}

// CreateMesh: Creates a new mesh, stores it and returns the mesh id
func (m *MeshManagerImpl) CreateMesh(devName string, port int) (string, error) {
	meshId, err := m.idGenerator.GetId()

	if err != nil {
		return "", err
	}

	nodeManager, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName: devName,
		Port:    port,
		Conf:    m.conf,
		Client:  m.Client,
		MeshId:  meshId,
	})

	if err != nil {
		return "", err
	}

	err = m.interfaceManipulator.CreateInterface(&wg.CreateInterfaceParams{
		IfName: devName,
		Port:   port,
	})

	if err != nil {
		return "", nil
	}

	m.Meshes[meshId] = nodeManager

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
	}

	return meshId, nil
}

type AddMeshParams struct {
	MeshId    string
	DevName   string
	WgPort    int
	MeshBytes []byte
}

// AddMesh: Add the mesh to the list of meshes
func (m *MeshManagerImpl) AddMesh(params *AddMeshParams) error {
	meshProvider, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName: params.DevName,
		Port:    params.WgPort,
		Conf:    m.conf,
		Client:  m.Client,
		MeshId:  params.MeshId,
	})

	if err != nil {
		return err
	}

	err = meshProvider.Load(params.MeshBytes)

	if err != nil {
		return err
	}

	m.Meshes[params.MeshId] = meshProvider

	return m.interfaceManipulator.CreateInterface(&wg.CreateInterfaceParams{
		IfName: params.DevName,
		Port:   params.WgPort,
	})
}

// HasChanges returns true if the mesh has changes
func (m *MeshManagerImpl) HasChanges(meshId string) bool {
	return m.Meshes[meshId].HasChanges()
}

// GetMesh returns the mesh with the given meshid
func (m *MeshManagerImpl) GetMesh(meshId string) MeshProvider {
	theMesh := m.Meshes[meshId]
	return theMesh
}

// EnableInterface: Enables the given WireGuard interface.
func (s *MeshManagerImpl) EnableInterface(meshId string) error {
	err := s.configApplyer.ApplyConfig()

	if err != nil {
		return err
	}

	meshNode, err := s.GetSelf(meshId)

	if err != nil {
		return err
	}

	mesh := s.GetMesh(meshId)

	if err != nil {
		return err
	}

	dev, err := mesh.GetDevice()

	if err != nil {
		return err
	}

	return s.interfaceManipulator.EnableInterface(dev.Name, meshNode.GetWgHost().String())
}

// GetPublicKey: Gets the public key of the WireGuard mesh
func (s *MeshManagerImpl) GetPublicKey(meshId string) (*wgtypes.Key, error) {
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

type AddSelfParams struct {
	// MeshId is the ID of the mesh to add this instance to
	MeshId string
	// WgPort is the WireGuard port to advertise
	WgPort int
	// Endpoint is the alias of the machine to send routable packets
	// to
	Endpoint string
}

// AddSelf adds this host to the mesh
func (s *MeshManagerImpl) AddSelf(params *AddSelfParams) error {
	pubKey, err := s.GetPublicKey(params.MeshId)

	if err != nil {
		return err
	}

	nodeIP, err := s.ipAllocator.GetIP(*pubKey, params.MeshId)

	if err != nil {
		return err
	}

	node := s.nodeFactory.Build(&MeshNodeFactoryParams{
		PublicKey: pubKey,
		NodeIP:    nodeIP,
		WgPort:    params.WgPort,
		Endpoint:  params.Endpoint,
	})

	s.Meshes[params.MeshId].AddNode(node)
	return s.RouteManager.UpdateRoutes()
}

// LeaveMesh leaves the mesh network
func (s *MeshManagerImpl) LeaveMesh(meshId string) error {
	_, exists := s.Meshes[meshId]

	if !exists {
		return errors.New(fmt.Sprintf("mesh %s does not exist", meshId))
	}

	// For now just delete the mesh with the ID.
	delete(s.Meshes, meshId)
	return nil
}

func (s *MeshManagerImpl) GetSelf(meshId string) (MeshNode, error) {
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

func (s *MeshManagerImpl) ApplyConfig() error {
	return s.configApplyer.ApplyConfig()
}

func (s *MeshManagerImpl) SetDescription(description string) error {
	for _, mesh := range s.Meshes {
		err := mesh.SetDescription(s.HostParameters.HostEndpoint, description)

		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateTimeStamp updates the timestamp of this node in all meshes
func (s *MeshManagerImpl) UpdateTimeStamp() error {
	for _, mesh := range s.Meshes {
		snapshot, err := mesh.GetMesh()

		if err != nil {
			return err
		}

		_, exists := snapshot.GetNodes()[s.HostParameters.HostEndpoint]

		if exists {
			err = mesh.UpdateTimeStamp(s.HostParameters.HostEndpoint)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *MeshManagerImpl) GetClient() *wgctrl.Client {
	return s.Client
}

func (s *MeshManagerImpl) GetMeshes() map[string]MeshProvider {
	return s.Meshes
}

// NewMeshManagerParams params required to create an instance of a mesh manager
type NewMeshManagerParams struct {
	Conf                 conf.WgMeshConfiguration
	Client               *wgctrl.Client
	MeshProvider         MeshProviderFactory
	NodeFactory          MeshNodeFactory
	IdGenerator          lib.IdGenerator
	IPAllocator          ip.IPAllocator
	InterfaceManipulator wg.WgInterfaceManipulator
	ConfigApplyer        MeshConfigApplyer
}

// Creates a new instance of a mesh manager with the given parameters
func NewMeshManager(params *NewMeshManagerParams) *MeshManagerImpl {
	hostParams := HostParameters{}

	switch params.Conf.Endpoint {
	case "":
		hostParams.HostEndpoint = fmt.Sprintf("%s:%s", lib.GetOutboundIP().String(), params.Conf.GrpcPort)
	default:
		hostParams.HostEndpoint = fmt.Sprintf("%s:%s", params.Conf.Endpoint, params.Conf.GrpcPort)
	}

	logging.Log.WriteInfof("Endpoint %s", hostParams.HostEndpoint)

	m := &MeshManagerImpl{
		Meshes:              make(map[string]MeshProvider),
		HostParameters:      &hostParams,
		meshProviderFactory: params.MeshProvider,
		nodeFactory:         params.NodeFactory,
		Client:              params.Client,
		conf:                &params.Conf,
	}

	m.configApplyer = params.ConfigApplyer
	m.RouteManager = NewRouteManager(m)
	m.idGenerator = params.IdGenerator
	m.ipAllocator = params.IPAllocator
	m.interfaceManipulator = params.InterfaceManipulator
	return m
}
