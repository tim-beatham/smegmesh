package mesh

import (
	"fmt"
	"net"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshNodeStub struct {
	hostEndpoint string
	publicKey    wgtypes.Key
	wgEndpoint   string
	wgHost       *net.IPNet
	timeStamp    int64
	routes       []string
	identifier   string
	description  string
}

// GetAlias implements MeshNode.
func (*MeshNodeStub) GetAlias() string {
	panic("unimplemented")
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

func (m *MeshNodeStub) GetRoutes() []string {
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

// SetAlias implements MeshProvider.
func (*MeshProviderStub) SetAlias(nodeId string, alias string) error {
	panic("unimplemented")
}

// RemoveRoutes implements MeshProvider.
func (*MeshProviderStub) RemoveRoutes(nodeId string, route ...string) error {
	panic("unimplemented")
}

// Prune implements MeshProvider.
func (*MeshProviderStub) Prune(pruneAmount int) error {
	return nil
}

// UpdateTimeStamp implements MeshProvider.
func (*MeshProviderStub) UpdateTimeStamp(nodeId string) error {
	return nil
}

func (s *MeshProviderStub) AddNode(node MeshNode) {
	s.snapshot.nodes[node.GetHostEndpoint()] = node
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

func (s *MeshProviderStub) AddRoutes(nodeId string, route ...string) error {
	return nil
}

func (s *MeshProviderStub) GetSyncer() MeshSyncer {
	return nil
}

func (s *MeshProviderStub) SetDescription(nodeId string, description string) error {
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
	Config *conf.WgMeshConfiguration
}

func (s *StubNodeFactory) Build(params *MeshNodeFactoryParams) MeshNode {
	_, wgHost, _ := net.ParseCIDR(fmt.Sprintf("%s/128", params.NodeIP.String()))

	return &MeshNodeStub{
		hostEndpoint: params.Endpoint,
		publicKey:    *params.PublicKey,
		wgEndpoint:   fmt.Sprintf("%s:%s", params.Endpoint, s.Config.GrpcPort),
		wgHost:       wgHost,
		timeStamp:    time.Now().Unix(),
		routes:       make([]string, 0),
		identifier:   "abc",
		description:  "A Mesh Node Stub",
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

// GetMonitor implements MeshManager.
func (*MeshManagerStub) GetMonitor() MeshMonitor {
	panic("unimplemented")
}

// SetAlias implements MeshManager.
func (*MeshManagerStub) SetAlias(alias string) error {
	panic("unimplemented")
}

// Close implements MeshManager.
func (*MeshManagerStub) Close() error {
	panic("unimplemented")
}

// Prune implements MeshManager.
func (*MeshManagerStub) Prune() error {
	return nil
}

func NewMeshManagerStub() MeshManager {
	return &MeshManagerStub{meshes: make(map[string]MeshProvider)}
}

func (m *MeshManagerStub) CreateMesh(devName string, port int) (string, error) {
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

func (m *MeshManagerStub) EnableInterface(meshId string) error {
	return nil
}

func (m *MeshManagerStub) GetPublicKey(meshId string) (*wgtypes.Key, error) {
	key, _ := wgtypes.GenerateKey()
	return &key, nil
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

func (m *MeshManagerStub) SetDescription(description string) error {
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