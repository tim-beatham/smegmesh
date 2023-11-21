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
	CreateMesh(port int) (string, error)
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
	SetAlias(alias string) error
	SetService(service string, value string) error
	RemoveService(service string) error
	UpdateTimeStamp() error
	GetClient() *wgctrl.Client
	GetMeshes() map[string]MeshProvider
	Prune() error
	Close() error
	GetMonitor() MeshMonitor
	GetNode(string, string) MeshNode
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
	Monitor              MeshMonitor
}

// RemoveService implements MeshManager.
func (m *MeshManagerImpl) RemoveService(service string) error {
	for _, mesh := range m.Meshes {
		err := mesh.RemoveService(m.HostParameters.HostEndpoint, service)

		if err != nil {
			return err
		}
	}

	return nil
}

// SetService implements MeshManager.
func (m *MeshManagerImpl) SetService(service string, value string) error {
	for _, mesh := range m.Meshes {
		err := mesh.AddService(m.HostParameters.HostEndpoint, service, value)

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MeshManagerImpl) GetNode(meshid, nodeId string) MeshNode {
	mesh, ok := m.Meshes[meshid]

	if !ok {
		return nil
	}

	node, err := mesh.GetNode(nodeId)

	if err != nil {
		return nil
	}

	return node
}

// GetMonitor implements MeshManager.
func (m *MeshManagerImpl) GetMonitor() MeshMonitor {
	return m.Monitor
}

// Prune implements MeshManager.
func (m *MeshManagerImpl) Prune() error {
	for _, mesh := range m.Meshes {
		err := mesh.Prune(m.conf.PruneTime)

		if err != nil {
			return err
		}
	}

	return nil
}

// CreateMesh: Creates a new mesh, stores it and returns the mesh id
func (m *MeshManagerImpl) CreateMesh(port int) (string, error) {
	meshId, err := m.idGenerator.GetId()

	var ifName string = ""

	if err != nil {
		return "", err
	}

	if !m.conf.StubWg {
		ifName, err = m.interfaceManipulator.CreateInterface(port)

		if err != nil {
			return "", fmt.Errorf("error creating mesh: %w", err)
		}
	}

	nodeManager, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName: ifName,
		Port:    port,
		Conf:    m.conf,
		Client:  m.Client,
		MeshId:  meshId,
	})

	if err != nil {
		return "", fmt.Errorf("error creating mesh: %w", err)
	}

	m.Meshes[meshId] = nodeManager
	return meshId, nil
}

type AddMeshParams struct {
	MeshId    string
	WgPort    int
	MeshBytes []byte
}

// AddMesh: Add the mesh to the list of meshes
func (m *MeshManagerImpl) AddMesh(params *AddMeshParams) error {
	var ifName string
	var err error

	if !m.conf.StubWg {
		ifName, err = m.interfaceManipulator.CreateInterface(params.WgPort)

		if err != nil {
			return err
		}
	}

	meshProvider, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName: ifName,
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
	return nil
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

	err = s.RouteManager.InstallRoutes()

	if err != nil {
		return err
	}

	return nil
}

// GetPublicKey: Gets the public key of the WireGuard mesh
func (s *MeshManagerImpl) GetPublicKey(meshId string) (*wgtypes.Key, error) {
	if s.conf.StubWg {
		zeroedKey := make([]byte, wgtypes.KeyLen)
		return (*wgtypes.Key)(zeroedKey), nil
	}

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
	Endpoint string
}

// AddSelf adds this host to the mesh
func (s *MeshManagerImpl) AddSelf(params *AddSelfParams) error {
	mesh := s.GetMesh(params.MeshId)

	if mesh == nil {
		return fmt.Errorf("addself: mesh %s does not exist", params.MeshId)
	}

	if params.WgPort == 0 && !s.conf.StubWg {
		device, err := mesh.GetDevice()

		if err != nil {
			return err
		}

		params.WgPort = device.ListenPort
	}

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
		Role:      s.conf.Role,
	})

	if !s.conf.StubWg {
		device, err := mesh.GetDevice()

		if err != nil {
			return fmt.Errorf("failed to get device %w", err)
		}

		err = s.interfaceManipulator.AddAddress(device.Name, fmt.Sprintf("%s/64", nodeIP))

		if err != nil {
			return fmt.Errorf("addSelf: failed to add address to dev %w", err)
		}
	}

	s.Meshes[params.MeshId].AddNode(node)
	return s.RouteManager.UpdateRoutes()
}

// LeaveMesh leaves the mesh network
func (s *MeshManagerImpl) LeaveMesh(meshId string) error {
	mesh, exists := s.Meshes[meshId]

	if !exists {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	err := s.RouteManager.RemoveRoutes(meshId)

	if err != nil {
		return err
	}

	if !s.conf.StubWg {
		device, e := mesh.GetDevice()

		if e != nil {
			return err
		}

		err = s.interfaceManipulator.RemoveInterface(device.Name)
	}

	delete(s.Meshes, meshId)
	return err
}

func (s *MeshManagerImpl) GetSelf(meshId string) (MeshNode, error) {
	meshInstance, ok := s.Meshes[meshId]

	if !ok {
		return nil, fmt.Errorf("mesh %s does not exist", meshId)
	}

	node, err := meshInstance.GetNode(s.HostParameters.HostEndpoint)

	if err != nil {
		return nil, errors.New("the node doesn't exist in the mesh")
	}

	return node, nil
}

func (s *MeshManagerImpl) ApplyConfig() error {
	if s.conf.StubWg {
		return nil
	}

	err := s.configApplyer.ApplyConfig()

	if err != nil {
		return err
	}

	return nil
}

func (s *MeshManagerImpl) SetDescription(description string) error {
	for _, mesh := range s.Meshes {
		if mesh.NodeExists(s.HostParameters.HostEndpoint) {
			err := mesh.SetDescription(s.HostParameters.HostEndpoint, description)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SetAlias implements MeshManager.
func (s *MeshManagerImpl) SetAlias(alias string) error {
	for _, mesh := range s.Meshes {
		if mesh.NodeExists(s.HostParameters.HostEndpoint) {
			err := mesh.SetAlias(s.HostParameters.HostEndpoint, alias)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateTimeStamp updates the timestamp of this node in all meshes
func (s *MeshManagerImpl) UpdateTimeStamp() error {
	for _, mesh := range s.Meshes {
		if mesh.NodeExists(s.HostParameters.HostEndpoint) {
			err := mesh.UpdateTimeStamp(s.HostParameters.HostEndpoint)

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

// Close the mesh manager
func (s *MeshManagerImpl) Close() error {
	if s.conf.StubWg {
		return nil
	}

	for _, mesh := range s.Meshes {
		dev, err := mesh.GetDevice()

		if err != nil {
			return err
		}

		err = s.interfaceManipulator.RemoveInterface(dev.Name)

		if err != nil {
			return err
		}
	}

	return nil
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
	RouteManager         RouteManager
}

// Creates a new instance of a mesh manager with the given parameters
func NewMeshManager(params *NewMeshManagerParams) MeshManager {
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
	m.RouteManager = params.RouteManager

	if m.RouteManager == nil {
		m.RouteManager = NewRouteManager(m)
	}

	m.idGenerator = params.IdGenerator
	m.ipAllocator = params.IPAllocator
	m.interfaceManipulator = params.InterfaceManipulator

	m.Monitor = NewMeshMonitor(m)

	aliasManager := NewAliasManager()
	m.Monitor.AddUpdateCallback(aliasManager.AddAliases)
	m.Monitor.AddRemoveCallback(aliasManager.RemoveAliases)
	return m
}
