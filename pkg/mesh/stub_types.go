package mesh

import (
	"fmt"
	"net"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshNodeStub struct {
	hostEndpoint string
	publicKey    wgtypes.Key
	wgEndpoint   string
	wgHost       *net.IPNet
	timeStamp    int64
	routes       []Route
	identifier   string
	description  string
	alias        string
	services     map[string]string
}

// GetType implements MeshNode.
func (*MeshNodeStub) GetType() conf.NodeType {
	return conf.PEER_ROLE
}

// GetServices implements MeshNode.
func (m *MeshNodeStub) GetServices() map[string]string {
	return m.services
}

// GetAlias implements MeshNode.
func (s *MeshNodeStub) GetAlias() string {
	return s.alias
}

func (m *MeshNodeStub) GetHostEndpoint() string {
	return m.hostEndpoint
}

func (m *MeshNodeStub) GetPublicKey() (wgtypes.Key, error) {
	return m.publicKey, nil
}

func (m *MeshNodeStub) GetWgEndpoint() string {
	return m.wgEndpoint
}

func (m *MeshNodeStub) GetWgHost() *net.IPNet {
	return m.wgHost
}

func (m *MeshNodeStub) GetTimeStamp() int64 {
	return m.timeStamp
}

func (m *MeshNodeStub) GetRoutes() []Route {
	return m.routes
}

func (m *MeshNodeStub) GetIdentifier() string {
	return m.identifier
}

func (m *MeshNodeStub) GetDescription() string {
	return m.description
}

type MeshSnapshotStub struct {
	nodes map[string]MeshNode
}

func (s *MeshSnapshotStub) GetNodes() map[string]MeshNode {
	return s.nodes
}

type MeshProviderStub struct {
	meshId   string
	snapshot *MeshSnapshotStub
}

// GetConfiguration implements MeshProvider.
func (*MeshProviderStub) GetConfiguration() *conf.WgConfiguration {
	advertiseRoutes := true
	advertiseDefaultRoute := true
	ipDiscovery := conf.PUBLIC_IP_DISCOVERY
	role := conf.PEER_ROLE

	return &conf.WgConfiguration{
		IPDiscovery:           &ipDiscovery,
		AdvertiseRoutes:       &advertiseRoutes,
		AdvertiseDefaultRoute: &advertiseDefaultRoute,
		Role:                  &role,
	}
}

// Mark implements MeshProvider.
func (*MeshProviderStub) Mark(nodeId string) {
}

// RemoveNode implements MeshProvider.
func (*MeshProviderStub) RemoveNode(nodeId string) error {
	return nil
}

func (*MeshProviderStub) GetRoutes(targetId string) (map[string]Route, error) {
	return nil, nil
}

// GetNodeIds implements MeshProvider.
func (*MeshProviderStub) GetPeers() []string {
	return make([]string, 0)
}

// GetNode implements MeshProvider.
func (m *MeshProviderStub) GetNode(nodeId string) (MeshNode, error) {
	return m.snapshot.nodes[nodeId], nil
}

// NodeExists implements MeshProvider.
func (m *MeshProviderStub) NodeExists(nodeId string) bool {
	return m.snapshot.nodes[nodeId] != nil
}

// AddService implements MeshProvider.
func (m *MeshProviderStub) AddService(nodeId string, key string, value string) error {
	node := (m.snapshot.nodes[nodeId]).(*MeshNodeStub)
	node.services[key] = value
	return nil
}

// RemoveService implements MeshProvider.
func (m *MeshProviderStub) RemoveService(nodeId string, key string) error {
	node := (m.snapshot.nodes[nodeId]).(*MeshNodeStub)
	delete(node.services, key)
	return nil
}

// SetAlias implements MeshProvider.
func (m *MeshProviderStub) SetAlias(nodeId string, alias string) error {
	node := (m.snapshot.nodes[nodeId]).(*MeshNodeStub)
	node.alias = alias
	return nil
}

// AddRoutes implements
func (m *MeshProviderStub) AddRoutes(nodeId string, route ...Route) error {
	node := (m.snapshot.nodes[nodeId]).(*MeshNodeStub)
	node.routes = append(node.routes, route...)
	return nil
}

// RemoveRoutes implements MeshProvider.
func (m *MeshProviderStub) RemoveRoutes(nodeId string, route ...Route) error {
	node := (m.snapshot.nodes[nodeId]).(*MeshNodeStub)

	newRoutes := lib.Filter(node.routes, func(r1 Route) bool {
		return !lib.Contains(route, func(r2 Route) bool {
			return RouteEqual(r1, r2)
		})
	})
	node.routes = newRoutes
	return nil
}

// Prune implements MeshProvider.
func (*MeshProviderStub) Prune() error {
	return nil
}

// UpdateTimeStamp implements MeshProvider.
func (m *MeshProviderStub) UpdateTimeStamp(nodeId string) error {
	node := (m.snapshot.nodes[nodeId]).(*MeshNodeStub)
	node.timeStamp = time.Now().Unix()
	return nil
}

func (s *MeshProviderStub) AddNode(node MeshNode) {
	pubKey, _ := node.GetPublicKey()
	s.snapshot.nodes[pubKey.String()] = node
}

func (s *MeshProviderStub) GetMesh() (MeshSnapshot, error) {
	return s.snapshot, nil
}

func (s *MeshProviderStub) GetMeshId() string {
	return s.meshId
}

func (s *MeshProviderStub) Save() []byte {
	return make([]byte, 0)
}

func (s *MeshProviderStub) Load(bytes []byte) error {
	return nil
}

func (s *MeshProviderStub) GetDevice() (*wgtypes.Device, error) {
	pubKey, _ := wgtypes.GenerateKey()
	return &wgtypes.Device{
		PublicKey: pubKey,
	}, nil
}

func (s *MeshProviderStub) SaveChanges() {}

func (s *MeshProviderStub) HasChanges() bool {
	return false
}

func (s *MeshProviderStub) GetSyncer() MeshSyncer {
	return nil
}

func (s *MeshProviderStub) SetDescription(nodeId string, description string) error {
	meshNode := (s.snapshot.nodes[nodeId]).(*MeshNodeStub)
	meshNode.description = description
	return nil
}

type StubMeshProviderFactory struct{}

func (s *StubMeshProviderFactory) CreateMesh(params *MeshProviderFactoryParams) (MeshProvider, error) {
	return &MeshProviderStub{
		meshId:   params.MeshId,
		snapshot: &MeshSnapshotStub{nodes: make(map[string]MeshNode)},
	}, nil
}

type StubNodeFactory struct {
	Config *conf.DaemonConfiguration
}

func (s *StubNodeFactory) Build(params *MeshNodeFactoryParams) MeshNode {
	_, wgHost, _ := net.ParseCIDR(fmt.Sprintf("%s/128", params.NodeIP.String()))

	return &MeshNodeStub{
		hostEndpoint: params.Endpoint,
		publicKey:    *params.PublicKey,
		wgEndpoint:   fmt.Sprintf("%s:%d", params.Endpoint, s.Config.GrpcPort),
		wgHost:       wgHost,
		timeStamp:    time.Now().Unix(),
		routes:       make([]Route, 0),
		identifier:   "abc",
		description:  "A Mesh Node Stub",
		services:     make(map[string]string),
	}
}

type MeshConfigApplyerStub struct{}

func (a *MeshConfigApplyerStub) ApplyConfig() error {
	return nil
}

func (a *MeshConfigApplyerStub) RemovePeers(meshId string) error {
	return nil
}

func (a *MeshConfigApplyerStub) SetMeshManager(manager MeshManager) {
}

type MeshManagerStub struct {
	meshes map[string]MeshProvider
}

// GetRouteManager implements MeshManager.
func (*MeshManagerStub) GetRouteManager() RouteManager {
	return nil
}

// GetNode implements MeshManager.
func (*MeshManagerStub) GetNode(meshId, nodeId string) MeshNode {
	return nil
}

// RemoveService implements MeshManager.
func (*MeshManagerStub) RemoveService(meshId, service string) error {
	return nil
}

// SetService implements MeshManager.
func (*MeshManagerStub) SetService(meshId, service, value string) error {
	return nil
}

// SetAlias implements MeshManager.
func (*MeshManagerStub) SetAlias(meshId, alias string) error {
	return nil
}

// Close implements MeshManager.
func (*MeshManagerStub) Close() error {
	return nil
}

// Prune implements MeshManager.
func (*MeshManagerStub) Prune() error {
	return nil
}

func NewMeshManagerStub() MeshManager {
	return &MeshManagerStub{meshes: make(map[string]MeshProvider)}
}

func (m *MeshManagerStub) CreateMesh(*CreateMeshParams) (string, error) {
	return "tim123", nil
}

func (m *MeshManagerStub) AddMesh(params *AddMeshParams) error {
	m.meshes[params.MeshId] = &MeshProviderStub{
		params.MeshId,
		&MeshSnapshotStub{nodes: make(map[string]MeshNode)},
	}

	return nil
}

func (m *MeshManagerStub) HasChanges(meshId string) bool {
	return false
}

func (m *MeshManagerStub) GetMesh(meshId string) MeshProvider {
	return &MeshProviderStub{
		meshId:   meshId,
		snapshot: &MeshSnapshotStub{nodes: make(map[string]MeshNode)}}
}

func (m *MeshManagerStub) GetPublicKey() *wgtypes.Key {
	key, _ := wgtypes.GenerateKey()
	return &key
}

func (m *MeshManagerStub) AddSelf(params *AddSelfParams) error {
	return nil
}

func (m *MeshManagerStub) GetSelf(meshId string) (MeshNode, error) {
	return nil, nil
}

func (m *MeshManagerStub) ApplyConfig() error {
	return nil
}

func (m *MeshManagerStub) SetDescription(meshId, description string) error {
	return nil
}

func (m *MeshManagerStub) UpdateTimeStamp() error {
	return nil
}

func (m *MeshManagerStub) GetClient() *wgctrl.Client {
	return nil
}

func (m *MeshManagerStub) GetMeshes() map[string]MeshProvider {
	return m.meshes
}

func (m *MeshManagerStub) LeaveMesh(meshId string) error {
	return nil
}
