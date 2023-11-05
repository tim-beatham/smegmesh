package crdt

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/automerge/automerge-go"
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
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

type NewCrdtNodeMangerParams struct {
	MeshId  string
	DevName string
	Port    int
	Conf    conf.WgMeshConfiguration
	Client  *wgctrl.Client
}

// NewCrdtNodeManager: Create a new crdt node manager
func NewCrdtNodeManager(params *NewCrdtNodeMangerParams) (*CrdtMeshManager, error) {
	var manager CrdtMeshManager
	manager.MeshId = params.MeshId
	manager.doc = automerge.New()
	manager.IfName = params.DevName
	manager.Client = params.Client
	manager.conf = &params.Conf
	return &manager, nil
}

// GetNode: returns a mesh node crdt.Close releases resources used by a Client.
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
		logging.Log.WriteInfof("Timestamp Updated for %s", nodeId)
	}

	return err
}

func (m *CrdtMeshManager) SetDescription(nodeId string, description string) error {
	node, err := m.doc.Path("nodes").Map().Get(nodeId)

	if err != nil {
		return err
	}

	if node.Kind() != automerge.KindMap {
		return fmt.Errorf("%s does not exist", nodeId)
	}

	err = node.Map().Set("description", description)

	if err == nil {
		logging.Log.WriteInfof("Description Updated for %s", nodeId)
	}

	return err
}

// AddRoutes: adds routes to the specific nodeId
func (m *CrdtMeshManager) AddRoutes(nodeId string, routes ...string) error {
	nodeVal, err := m.doc.Path("nodes").Map().Get(nodeId)
	logging.Log.WriteInfof("Adding route to %s", nodeId)

	if err != nil {
		return err
	}

	if nodeVal.Kind() != automerge.KindMap {
		return fmt.Errorf("node does not exist")
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
	return m.WgEndpoint
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

func (m *MeshNodeCrdt) GetDescription() string {
	return m.Description
}

func (m *MeshNodeCrdt) GetIdentifier() string {
	ipv6 := m.WgHost[:len(m.WgHost)-4]

	constituents := strings.Split(ipv6, ":")
	logging.Log.WriteInfof(ipv6)
	constituents = constituents[4:]
	return strings.Join(constituents, ":")
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
			Description:  node.Description,
		}
	}

	return nodes
}
