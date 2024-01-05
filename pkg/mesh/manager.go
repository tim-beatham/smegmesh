package mesh

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/tim-beatham/smegmesh/pkg/cmd"
	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/ip"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// MeshManager: abstracts maanging meshes, including installing the WireGuard configuration
// to the device, and adding and removing nodes
type MeshManager interface {
	CreateMesh(params *CreateMeshParams) (string, error)
	AddMesh(params *AddMeshParams) error
	HasChanges(meshid string) bool
	GetMesh(meshId string) MeshProvider
	GetPublicKey() *wgtypes.Key
	AddSelf(params *AddSelfParams) error
	LeaveMesh(meshId string) error
	GetSelf(meshId string) (MeshNode, error)
	ApplyConfig() error
	SetDescription(meshId, description string) error
	SetAlias(meshId, alias string) error
	SetService(meshId, service, value string) error
	RemoveService(meshId, service string) error
	UpdateTimeStamp() error
	GetClient() *wgctrl.Client
	GetMeshes() map[string]MeshProvider
	Close() error
	GetNode(string, string) MeshNode
	GetRouteManager() RouteManager
}

type MeshManagerImpl struct {
	meshLock             sync.RWMutex
	meshes               map[string]MeshProvider
	RouteManager         RouteManager
	Client               *wgctrl.Client
	HostParameters       *HostParameters
	conf                 *conf.DaemonConfiguration
	meshProviderFactory  MeshProviderFactory
	nodeFactory          MeshNodeFactory
	configApplyer        MeshConfigApplyer
	idGenerator          lib.IdGenerator
	ipAllocator          ip.IPAllocator
	interfaceManipulator wg.WgInterfaceManipulator
	cmdRunner            cmd.CmdRunner
	OnDelete             func(MeshProvider)
}

func (m *MeshManagerImpl) GetRouteManager() RouteManager {
	return m.RouteManager
}

// RemoveService: remove a service from the given mesh.
func (m *MeshManagerImpl) RemoveService(meshId, service string) error {
	mesh := m.GetMesh(meshId)

	if mesh == nil {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	if !mesh.NodeExists(m.HostParameters.GetPublicKey()) {
		return fmt.Errorf("node %s does not exist in the mesh", meshId)
	}

	return mesh.RemoveService(m.HostParameters.GetPublicKey(), service)
}

// SetService: add a service to the given mesh
func (m *MeshManagerImpl) SetService(meshId, service, value string) error {
	mesh := m.GetMesh(meshId)

	if mesh == nil {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	if !mesh.NodeExists(m.HostParameters.GetPublicKey()) {
		return fmt.Errorf("node %s does not exist in the mesh", meshId)
	}

	return mesh.AddService(m.HostParameters.GetPublicKey(), service, value)
}

// GetNode: gets the node with given id in the mesh network
func (m *MeshManagerImpl) GetNode(meshid, nodeId string) MeshNode {
	mesh, ok := m.meshes[meshid]

	if !ok {
		return nil
	}

	node, err := mesh.GetNode(nodeId)

	if err != nil {
		return nil
	}

	return node
}

// CreateMeshParams contains the parameters required to create a mesh
type CreateMeshParams struct {
	Port int
	Conf *conf.WgConfiguration
}

// getConf: gets the new configuration with the base configuration overriden
// from the recent
func (m *MeshManagerImpl) getConf(override *conf.WgConfiguration) (*conf.WgConfiguration, error) {
	meshConfiguration := m.conf.BaseConfiguration

	if override != nil {
		newConf, err := conf.MergeMeshConfiguration(meshConfiguration, *override)

		if err != nil {
			return nil, err
		}

		meshConfiguration = newConf
	}

	return &meshConfiguration, nil
}

// CreateMesh: Creates a new mesh, stores it and returns the mesh id
func (m *MeshManagerImpl) CreateMesh(args *CreateMeshParams) (string, error) {
	meshConfiguration, err := m.getConf(args.Conf)

	if err != nil {
		return "", err
	}

	if *meshConfiguration.Role == conf.CLIENT_ROLE {
		return "", fmt.Errorf("cannot create mesh as a client")
	}

	meshId, err := m.idGenerator.GetId()

	var ifName string = ""

	if err != nil {
		return "", err
	}

	m.cmdRunner.RunCommands(m.conf.BaseConfiguration.PreUp...)

	if !m.conf.StubWg {
		ifName, err = m.interfaceManipulator.CreateInterface(args.Port, m.HostParameters.PrivateKey)

		if err != nil {
			return "", fmt.Errorf("error creating mesh: %w", err)
		}
	}

	nodeManager, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName:    ifName,
		Port:       args.Port,
		Conf:       meshConfiguration,
		Client:     m.Client,
		MeshId:     meshId,
		DaemonConf: m.conf,
		NodeID:     m.HostParameters.GetPublicKey(),
	})

	if err != nil {
		return "", fmt.Errorf("error creating mesh: %w", err)
	}

	m.meshLock.Lock()
	m.meshes[meshId] = nodeManager
	m.meshLock.Unlock()

	m.cmdRunner.RunCommands(m.conf.BaseConfiguration.PostUp...)

	return meshId, nil
}

type AddMeshParams struct {
	MeshId    string
	WgPort    int
	MeshBytes []byte
	Conf      *conf.WgConfiguration
}

// AddMesh: Add a new mesh network to the list of addresses
func (m *MeshManagerImpl) AddMesh(params *AddMeshParams) error {
	var ifName string
	var err error

	meshConfiguration, err := m.getConf(params.Conf)

	if err != nil {
		return err
	}

	m.cmdRunner.RunCommands(meshConfiguration.PreUp...)

	if !m.conf.StubWg {
		ifName, err = m.interfaceManipulator.CreateInterface(params.WgPort, m.HostParameters.PrivateKey)

		if err != nil {
			return err
		}
	}

	meshProvider, err := m.meshProviderFactory.CreateMesh(&MeshProviderFactoryParams{
		DevName:    ifName,
		Port:       params.WgPort,
		Conf:       meshConfiguration,
		Client:     m.Client,
		MeshId:     params.MeshId,
		DaemonConf: m.conf,
		NodeID:     m.HostParameters.GetPublicKey(),
	})

	m.cmdRunner.RunCommands(meshConfiguration.PostUp...)

	if err != nil {
		return err
	}

	err = meshProvider.Load(params.MeshBytes)

	if err != nil {
		return err
	}

	m.meshLock.Lock()
	m.meshes[params.MeshId] = meshProvider
	m.meshLock.Unlock()
	return nil
}

// HasChanges: returns true if the mesh has changes
func (m *MeshManagerImpl) HasChanges(meshId string) bool {
	return m.meshes[meshId].HasChanges()
}

// GetMesh: returns the mesh with the given meshid
func (m *MeshManagerImpl) GetMesh(meshId string) MeshProvider {
	theMesh := m.meshes[meshId]
	return theMesh
}

// GetPublicKey: Gets the public key of the WireGuard mesh
func (s *MeshManagerImpl) GetPublicKey() *wgtypes.Key {
	key := s.HostParameters.PrivateKey.PublicKey()
	return &key
}

// AddSelfParams: parameters required to add yourself to a mesh
// network
type AddSelfParams struct {
	// MeshId is the ID of the mesh to add this instance to
	MeshId string
	// WgPort is the WireGuard port to advertise
	WgPort int
	// Endpoint is the alias of the machine to send routable packets
	Endpoint string
}

// AddSelf: adds this host to the mesh
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

	pubKey := s.HostParameters.PrivateKey.PublicKey()

	collisionCount := uint8(0)

	var nodeIP net.IP

	// Perform Duplicate Address Detection with the nodes
	// that are already in the network
	for {
		generatedIP, err := s.ipAllocator.GetIP(pubKey, params.MeshId, collisionCount)

		if err != nil {
			return err
		}

		snapshot, err := mesh.GetMesh()

		if err != nil {
			return err
		}

		proposition := func(node MeshNode) bool {
			ipNet := node.GetWgHost()
			return ipNet.IP.Equal(nodeIP)
		}

		if lib.Contains(lib.MapValues(snapshot.GetNodes()), proposition) {
			collisionCount++
		} else {
			nodeIP = generatedIP
			break
		}
	}

	node := s.nodeFactory.Build(&MeshNodeFactoryParams{
		PublicKey:  &pubKey,
		NodeIP:     nodeIP,
		WgPort:     params.WgPort,
		Endpoint:   params.Endpoint,
		MeshConfig: mesh.GetConfiguration(),
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

	s.meshes[params.MeshId].AddNode(node)
	return nil
}

// LeaveMesh: leaves the mesh network and force a synchronsiation
func (s *MeshManagerImpl) LeaveMesh(meshId string) error {
	mesh := s.GetMesh(meshId)

	if mesh == nil {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	err := mesh.RemoveNode(s.HostParameters.GetPublicKey())

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
	}

	if s.OnDelete != nil {
		s.OnDelete(mesh)
	}

	s.meshLock.Lock()
	delete(s.meshes, meshId)
	s.meshLock.Unlock()

	s.cmdRunner.RunCommands(s.conf.BaseConfiguration.PreDown...)

	if !s.conf.StubWg {
		device, err := mesh.GetDevice()

		if err != nil {
			return err
		}

		err = s.interfaceManipulator.RemoveInterface(device.Name)

		if err != nil {
			return err
		}
	}

	s.cmdRunner.RunCommands(s.conf.BaseConfiguration.PostDown...)
	return err
}

func (s *MeshManagerImpl) GetSelf(meshId string) (MeshNode, error) {
	meshInstance, ok := s.meshes[meshId]

	if !ok {
		return nil, fmt.Errorf("mesh %s does not exist", meshId)
	}

	node, err := meshInstance.GetNode(s.HostParameters.GetPublicKey())

	if err != nil {
		return nil, errors.New("the node doesn't exist in the mesh")
	}

	return node, nil
}

// ApplyConfig: applies the WireGuard configuration
// adds routes to the RIB and so forth.
func (s *MeshManagerImpl) ApplyConfig() error {
	if s.conf.StubWg {
		return nil
	}
	return s.configApplyer.ApplyConfig()
}

func (s *MeshManagerImpl) SetDescription(meshId, description string) error {
	mesh := s.GetMesh(meshId)

	if mesh == nil {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	if !mesh.NodeExists(s.HostParameters.GetPublicKey()) {
		return fmt.Errorf("node %s does not exist in the mesh", meshId)
	}

	return mesh.SetDescription(s.HostParameters.GetPublicKey(), description)
}

// SetAlias sets the alias of the node for the given meshid
func (s *MeshManagerImpl) SetAlias(meshId, alias string) error {
	mesh := s.GetMesh(meshId)

	if mesh == nil {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	if !mesh.NodeExists(s.HostParameters.GetPublicKey()) {
		return fmt.Errorf("node %s does not exist in the mesh", meshId)
	}

	return mesh.SetAlias(s.HostParameters.GetPublicKey(), alias)
}

// UpdateTimeStamp: updates the timestamp of this node in all meshes
// essentially performs heartbeat if the node is the leader
func (s *MeshManagerImpl) UpdateTimeStamp() error {
	meshes := s.GetMeshes()
	for _, mesh := range meshes {
		if mesh.NodeExists(s.HostParameters.GetPublicKey()) {
			err := mesh.UpdateTimeStamp(s.HostParameters.GetPublicKey())

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

// GetMeshes: get all meshes the node is part of
func (s *MeshManagerImpl) GetMeshes() map[string]MeshProvider {
	meshes := make(map[string]MeshProvider)

	// GetMesh: copies the map of meshes to a new map
	// to prevent a whole range of concurrency issues
	// due to iteration and modification
	s.meshLock.RLock()

	for id, mesh := range s.meshes {
		meshes[id] = mesh
	}

	s.meshLock.RUnlock()
	return meshes
}

// Close: close the mesh manager
func (s *MeshManagerImpl) Close() error {
	if s.conf.StubWg {
		return nil
	}

	for _, mesh := range s.meshes {
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

// NewMeshManagerParams: params required to create an instance of a mesh manager
type NewMeshManagerParams struct {
	Conf                 conf.DaemonConfiguration
	Client               *wgctrl.Client
	MeshProvider         MeshProviderFactory
	NodeFactory          MeshNodeFactory
	IdGenerator          lib.IdGenerator
	IPAllocator          ip.IPAllocator
	InterfaceManipulator wg.WgInterfaceManipulator
	ConfigApplyer        MeshConfigApplyer
	RouteManager         RouteManager
	CommandRunner        cmd.CmdRunner
	OnDelete             func(MeshProvider)
}

// NewMeshManager: Creates a new instance of a mesh manager with the given parameters
func NewMeshManager(params *NewMeshManagerParams) MeshManager {
	privateKey, _ := wgtypes.GeneratePrivateKey()
	hostParams := HostParameters{
		PrivateKey: &privateKey,
	}

	m := &MeshManagerImpl{
		meshes:              make(map[string]MeshProvider),
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

	if params.CommandRunner == nil {
		m.cmdRunner = &cmd.UnixCmdRunner{}
	}

	m.idGenerator = params.IdGenerator
	m.ipAllocator = params.IPAllocator
	m.interfaceManipulator = params.InterfaceManipulator

	m.OnDelete = params.OnDelete
	return m
}
