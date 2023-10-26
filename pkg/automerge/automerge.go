package crdt

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/automerge/automerge-go"
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// CrdtMeshManager manages nodes in the crdt mesh
type CrdtMeshManager struct {
	MeshId   string
	IfName   string
	NodeId   string
	Client   *wgctrl.Client
	doc      *automerge.Doc
	LastHash automerge.ChangeHash
	conf     *conf.WgMeshConfiguration
}

func (c *CrdtMeshManager) AddNode(node mesh.MeshNode) {
	crdt, ok := node.(*MeshNodeCrdt)

	if !ok {
		panic("node must be of type *MeshNodeCrdt")
	}

	crdt.Timestamp = time.Now().Unix()
	c.doc.Path("nodes").Map().Set(crdt.HostEndpoint, crdt)
	nodeVal, _ := c.doc.Path("nodes").Map().Get(crdt.HostEndpoint)
	nodeVal.Map().Set("routes", automerge.NewMap())
}

func (c *CrdtMeshManager) ApplyWg() error {
	// snapshot, err := c.GetMesh()

	// if err != nil {
	// return err
	// }

	// c.updateWgConf(c.IfName, snapshot.GetNodes(), *c.Client)
	return nil
}

// GetMesh(): Converts the document into a struct
func (c *CrdtMeshManager) GetMesh() (mesh.MeshSnapshot, error) {
	return automerge.As[*MeshCrdt](c.doc.Root())
}

// GetMeshId returns the meshid of the mesh
func (c *CrdtMeshManager) GetMeshId() string {
	return c.MeshId
}

// Save: Save an entire mesh network
func (c *CrdtMeshManager) Save() []byte {
	return c.doc.Save()
}

// Load: Load an entire mesh network
func (c *CrdtMeshManager) Load(bytes []byte) error {
	doc, err := automerge.Load(bytes)

	if err != nil {
		return err
	}
	c.doc = doc
	return nil
}

// NewCrdtNodeManager: Create a new crdt node manager
func NewCrdtNodeManager(meshId, devName string, port int, conf conf.WgMeshConfiguration, client *wgctrl.Client) (*CrdtMeshManager, error) {
	var manager CrdtMeshManager
	manager.MeshId = meshId
	manager.doc = automerge.New()
	manager.IfName = devName
	manager.Client = client
	manager.conf = &conf

	err := wg.CreateWgInterface(client, devName, port)

	if err != nil {
		return nil, err
	}

	return &manager, nil
}

func (m *CrdtMeshManager) convertMeshNode(node MeshNodeCrdt) (*wgtypes.PeerConfig, error) {
	peerEndpoint, err := net.ResolveUDPAddr("udp", node.WgEndpoint)

	if err != nil {
		return nil, err
	}

	peerPublic, err := wgtypes.ParseKey(node.PublicKey)

	if err != nil {
		return nil, err
	}

	allowedIps := make([]net.IPNet, 1)
	_, ipnet, err := net.ParseCIDR(node.WgHost)

	if err != nil {
		return nil, err
	}

	allowedIps[0] = *ipnet

	for route, _ := range node.Routes {
		_, ipnet, _ := net.ParseCIDR(route)
		allowedIps = append(allowedIps, *ipnet)
	}

	peerConfig := wgtypes.PeerConfig{
		PublicKey:  peerPublic,
		Remove:     m.HasFailed(node.HostEndpoint),
		Endpoint:   peerEndpoint,
		AllowedIPs: allowedIps,
	}

	return &peerConfig, nil
}

func (c *CrdtMeshManager) removeNode(endpoint string) error {
	err := c.doc.Path("nodes").Map().Delete(endpoint)

	if err != nil {
		return err
	}

	return nil
}

// GetNode: returns a mesh node crdt.
func (m *CrdtMeshManager) GetNode(endpoint string) (*MeshNodeCrdt, error) {
	node, err := m.doc.Path("nodes").Map().Get(endpoint)

	if err != nil {
		return nil, err
	}

	meshNode, err := automerge.As[*MeshNodeCrdt](node)

	if err != nil {
		return nil, err
	}

	return meshNode, nil
}

func (m *CrdtMeshManager) Length() int {
	return m.doc.Path("nodes").Map().Len()
}

func (m *CrdtMeshManager) GetDevice() (*wgtypes.Device, error) {
	dev, err := m.Client.Device(m.IfName)

	if err != nil {
		return nil, err
	}

	return dev, nil
}

// HasChanges returns true if we have changes since the last time we synced
func (m *CrdtMeshManager) HasChanges() bool {
	changes, err := m.doc.Changes(m.LastHash)

	logging.Log.WriteInfof("Changes %s", m.LastHash.String())

	if err != nil {
		return false
	}

	logging.Log.WriteInfof("Changes length %d", len(changes))
	return len(changes) > 0
}

func (m *CrdtMeshManager) HasFailed(endpoint string) bool {
	return false
}

func (m *CrdtMeshManager) SaveChanges() {
	hashes := m.doc.Heads()
	hash := hashes[len(hashes)-1]

	logging.Log.WriteInfof("Saved Hash %s", hash.String())
	m.LastHash = hash
}

func (m *CrdtMeshManager) UpdateTimeStamp(nodeId string) error {
	node, err := m.doc.Path("nodes").Map().Get(nodeId)

	if err != nil {
		return err
	}

	if node.Kind() != automerge.KindMap {
		return errors.New("node is not a map")
	}

	err = node.Map().Set("timestamp", time.Now().Unix())

	if err == nil {
		logging.Log.WriteInfof("Timestamp Updated for %s", m.MeshId)
	}

	return err
}

// AddRoutes: adds routes to the specific nodeId
func (m *CrdtMeshManager) AddRoutes(nodeId string, routes ...string) error {
	nodeVal, err := m.doc.Path("nodes").Map().Get(nodeId)

	if err != nil {
		return err
	}

	routeMap, err := nodeVal.Map().Get("routes")

	if err != nil {
		return err
	}

	for _, route := range routes {
		err = routeMap.Map().Set(route, struct{}{})

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *CrdtMeshManager) updateWgConf(devName string, nodes map[string]MeshNodeCrdt, client wgctrl.Client) error {
	peerConfigs := make([]wgtypes.PeerConfig, len(nodes))

	var count int = 0

	for _, n := range nodes {
		peer, err := m.convertMeshNode(n)

		if err != nil {
			return err
		}

		peerConfigs[count] = *peer
		count++
	}

	cfg := wgtypes.Config{
		Peers:        peerConfigs,
		ReplacePeers: true,
	}

	client.ConfigureDevice(devName, cfg)
	return nil
}

func (m *CrdtMeshManager) GetSyncer() mesh.MeshSyncer {
	return NewAutomergeSync(m)
}

func (m1 *MeshNodeCrdt) Compare(m2 *MeshNodeCrdt) int {
	return strings.Compare(m1.PublicKey, m2.PublicKey)
}

func (m *MeshNodeCrdt) GetHostEndpoint() string {
	return m.HostEndpoint
}

func (m *MeshNodeCrdt) GetPublicKey() (wgtypes.Key, error) {
	return wgtypes.ParseKey(m.PublicKey)
}

func (m *MeshNodeCrdt) GetWgEndpoint() string {
	return m.HostEndpoint
}

func (m *MeshNodeCrdt) GetWgHost() *net.IPNet {
	_, ipnet, err := net.ParseCIDR(m.WgHost)

	if err != nil {
		logging.Log.WriteErrorf("Cannot parse WgHost %s", err.Error())
		return nil
	}

	return ipnet
}

func (m *MeshNodeCrdt) GetTimeStamp() int64 {
	return m.Timestamp
}

func (m *MeshNodeCrdt) GetRoutes() []string {
	return lib.MapKeys(m.Routes)
}

func (m *MeshCrdt) GetNodes() map[string]mesh.MeshNode {
	nodes := make(map[string]mesh.MeshNode)

	for _, node := range m.Nodes {
		nodes[node.HostEndpoint] = &MeshNodeCrdt{
			HostEndpoint: node.HostEndpoint,
			WgEndpoint:   node.WgEndpoint,
			PublicKey:    node.PublicKey,
			WgHost:       node.WgHost,
			Timestamp:    node.Timestamp,
			Routes:       node.Routes,
		}
	}

	return nodes
}
