package crdt

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/automerge/automerge-go"
	"github.com/tim-beatham/wgmesh/pkg/conf"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// CrdtNodeManager manages nodes in the crdt mesh
type CrdtNodeManager struct {
	MeshId   string
	IfName   string
	NodeId   string
	Client   *wgctrl.Client
	doc      *automerge.Doc
	LastHash automerge.ChangeHash
	conf     *conf.WgMeshConfiguration
}

const maxFails = 5

func (c *CrdtNodeManager) AddNode(crdt MeshNodeCrdt) {
	crdt.FailedMap = automerge.NewMap()
	crdt.Timestamp = time.Now().Unix()
	c.doc.Path("nodes").Map().Set(crdt.HostEndpoint, crdt)
	nodeVal, _ := c.doc.Path("nodes").Map().Get(crdt.HostEndpoint)
	nodeVal.Map().Set("routes", automerge.NewMap())
}

func (c *CrdtNodeManager) ApplyWg() error {
	snapshot, err := c.GetCrdt()

	if err != nil {
		return err
	}

	c.updateWgConf(c.IfName, snapshot.Nodes, *c.Client)
	return nil
}

// GetCrdt(): Converts the document into a struct
func (c *CrdtNodeManager) GetCrdt() (*MeshCrdt, error) {
	return automerge.As[*MeshCrdt](c.doc.Root())
}

// Load: Load an entire mesh network
func (c *CrdtNodeManager) Load(bytes []byte) error {
	doc, err := automerge.Load(bytes)

	if err != nil {
		return err
	}

	c.doc = doc
	return nil
}

// Save: Save an entire mesh network
func (c *CrdtNodeManager) Save() []byte {
	return c.doc.Save()
}

// NewCrdtNodeManager: Create a new crdt node manager
func NewCrdtNodeManager(meshId, hostId, devName string, port int, conf conf.WgMeshConfiguration, client *wgctrl.Client) (*CrdtNodeManager, error) {
	var manager CrdtNodeManager
	manager.MeshId = meshId
	manager.doc = automerge.New()
	manager.IfName = devName
	manager.Client = client
	manager.NodeId = hostId
	manager.conf = &conf

	err := wg.CreateWgInterface(client, devName, port)

	if err != nil {
		return nil, err
	}

	return &manager, nil
}

func (m *CrdtNodeManager) convertMeshNode(node MeshNodeCrdt) (*wgtypes.PeerConfig, error) {
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

func (m1 *MeshNodeCrdt) Compare(m2 *MeshNodeCrdt) int {
	return strings.Compare(m1.PublicKey, m2.PublicKey)
}

func (c *CrdtNodeManager) changeFailedCount(meshId, endpoint string, incAmount int64) error {
	node, err := c.doc.Path("nodes").Map().Get(endpoint)

	if err != nil {
		return err
	}

	counterMap, err := node.Map().Get("failedMap")

	if counterMap.Kind() == automerge.KindVoid {
		return errors.New("Something went wrong map does not exist")
	}

	counter, _ := counterMap.Map().Get(c.NodeId)

	if counter.Kind() == automerge.KindVoid {
		err = counterMap.Map().Set(c.NodeId, incAmount)
	} else {
		if counter.Int64()+incAmount < 0 {
			return nil
		}

		err = counterMap.Map().Set(c.NodeId, counter.Int64()+1)
	}

	return err
}

// Increment failed count increments the number of times we have attempted
// to contact the node and it's failed
func (c *CrdtNodeManager) IncrementFailedCount(endpoint string) error {
	return c.changeFailedCount(c.MeshId, endpoint, +1)
}

func (c *CrdtNodeManager) removeNode(endpoint string) error {
	err := c.doc.Path("nodes").Map().Delete(endpoint)

	if err != nil {
		return err
	}

	return nil
}

// Decrement failed count decrements the number of times we have attempted to
// contact the node and it's failed
func (c *CrdtNodeManager) DecrementFailedCount(endpoint string) error {
	return c.changeFailedCount(c.MeshId, endpoint, -1)
}

// GetNode: returns a mesh node crdt.
func (m *CrdtNodeManager) GetNode(endpoint string) (*MeshNodeCrdt, error) {
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

func (m *CrdtNodeManager) Length() int {
	return m.doc.Path("nodes").Map().Len()
}

func (m *CrdtNodeManager) GetDevice() (*wgtypes.Device, error) {
	dev, err := m.Client.Device(m.IfName)

	if err != nil {
		return nil, err
	}

	return dev, nil
}

// HasChanges returns true if we have changes since the last time we synced
func (m *CrdtNodeManager) HasChanges() bool {
	changes, err := m.doc.Changes(m.LastHash)

	logging.Log.WriteInfof("Changes %s", m.LastHash.String())

	if err != nil {
		return false
	}

	logging.Log.WriteInfof("Changes length %d", len(changes))
	return len(changes) > 0
}

func (m *CrdtNodeManager) HasFailed(endpoint string) bool {
	node, err := m.GetNode(endpoint)

	if err != nil {
		logging.Log.WriteErrorf("Cannot get node node: %s\n", endpoint)
		return true
	}

	values, err := node.FailedMap.Values()

	if err != nil {
		return true
	}

	countFailed := 0

	for _, value := range values {
		count := value.Int64()

		if count >= 1 {
			countFailed++
		}
	}

	return countFailed >= 4
}

func (m *CrdtNodeManager) SaveChanges() {
	hashes := m.doc.Heads()
	hash := hashes[len(hashes)-1]

	logging.Log.WriteInfof("Saved Hash %s", hash.String())
	m.LastHash = hash
}

func (m *CrdtNodeManager) UpdateTimeStamp() error {
	node, err := m.doc.Path("nodes").Map().Get(m.NodeId)

	if err != nil {
		return err
	}

	err = node.Map().Set("timestamp", time.Now().Unix())

	if err == nil {
		logging.Log.WriteInfof("Timestamp Updated for %s", m.MeshId)
	}

	return err
}

func (m *CrdtNodeManager) updateWgConf(devName string, nodes map[string]MeshNodeCrdt, client wgctrl.Client) error {
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

// AddRoutes: adds routes to the specific nodeId
func (m *CrdtNodeManager) AddRoutes(routes ...string) error {
	nodeVal, err := m.doc.Path("nodes").Map().Get(m.NodeId)

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

func (m *CrdtNodeManager) GetSyncer() *AutomergeSync {
	return NewAutomergeSync(m)
}

func (n *MeshNodeCrdt) GetEscapedIP() string {
	return fmt.Sprintf("\"%s\"", n.WgHost)
}
