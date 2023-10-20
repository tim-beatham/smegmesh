package crdt

import (
	"net"
	"strings"

	"github.com/automerge/automerge-go"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// CrdtNodeManager manages nodes in the crdt mesh
type CrdtNodeManager struct {
	MeshId string
	IfName string
	Client *wgctrl.Client
	doc    *automerge.Doc
}

const maxFails = 5

func (c *CrdtNodeManager) AddNode(crdt MeshNodeCrdt) {
	crdt.FailedCount = automerge.NewCounter(0)
	c.doc.Path("nodes").Map().Set(crdt.HostEndpoint, crdt)

}

func (c *CrdtNodeManager) applyWg() error {
	snapshot, err := c.GetCrdt()

	if err != nil {
		return err
	}

	updateWgConf(c.IfName, snapshot.Nodes, *c.Client)
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
	c.applyWg()
	return nil
}

// Save: Save an entire mesh network
func (c *CrdtNodeManager) Save() []byte {
	return c.doc.Save()
}

func (c *CrdtNodeManager) LoadChanges(changes []byte) error {
	err := c.doc.LoadIncremental(changes)

	if err != nil {
		return err
	}

	return c.applyWg()
}

func (c *CrdtNodeManager) SaveChanges() []byte {
	return c.doc.SaveIncremental()
}

// NewCrdtNodeManager: Create a new crdt node manager
func NewCrdtNodeManager(meshId, devName string, client *wgctrl.Client) *CrdtNodeManager {
	var manager CrdtNodeManager
	manager.MeshId = meshId
	manager.doc = automerge.New()
	manager.IfName = devName
	manager.Client = client
	return &manager
}

func convertMeshNode(node MeshNodeCrdt) (*wgtypes.PeerConfig, error) {
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

	peerConfig := wgtypes.PeerConfig{
		PublicKey:  peerPublic,
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

	counter, err := node.Map().Get("failedCount")

	if err != nil {
		return err
	}

	err = counter.Counter().Inc(incAmount)
	return err
}

// Increment failed count increments the number of times we have attempted
// to contact the node and it's failed
func (c *CrdtNodeManager) IncrementFailedCount(endpoint string) error {
	snapshot, err := c.GetCrdt()

	if err != nil {
		return err
	}

	count, err := snapshot.Nodes[endpoint].FailedCount.Get()

	if err != nil {
		return err
	}

	if count >= maxFails {
		c.removeNode(endpoint)
		logging.InfoLog.Printf("Node %s removed from mesh %s", endpoint, c.MeshId)
		return nil
	}

	if err != nil {
		return err
	}

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
	snapshot, err := c.GetCrdt()

	if err != nil {
		return err
	}

	count, err := snapshot.Nodes[endpoint].FailedCount.Get()

	if err != nil {
		return err
	}

	if count < 0 {
		return nil
	}

	return c.changeFailedCount(c.MeshId, endpoint, -1)
}

func updateWgConf(devName string, nodes map[string]MeshNodeCrdt, client wgctrl.Client) error {
	peerConfigs := make([]wgtypes.PeerConfig, len(nodes))

	var count int = 0

	for _, n := range nodes {
		peer, err := convertMeshNode(n)
		logging.InfoLog.Println(n.HostEndpoint)

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
