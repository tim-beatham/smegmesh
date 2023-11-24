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
	MeshId        string
	IfName        string
	Client        *wgctrl.Client
	doc           *automerge.Doc
	LastHash      automerge.ChangeHash
	conf          *conf.WgMeshConfiguration
	cache         *MeshCrdt
	lastCacheHash automerge.ChangeHash
}

func (c *CrdtMeshManager) AddNode(node mesh.MeshNode) {
	crdt, ok := node.(*MeshNodeCrdt)

	if !ok {
		panic("node must be of type *MeshNodeCrdt")
	}

	crdt.Routes = make(map[string]interface{})
	crdt.Services = make(map[string]string)
	crdt.Timestamp = time.Now().Unix()

	c.doc.Path("nodes").Map().Set(crdt.PublicKey, crdt)
}

func (c *CrdtMeshManager) isPeer(nodeId string) bool {
	node, err := c.doc.Path("nodes").Map().Get(nodeId)

	if err != nil || node.Kind() != automerge.KindMap {
		return false
	}

	nodeType, err := node.Map().Get("type")

	if err != nil || nodeType.Kind() != automerge.KindStr {
		return false
	}

	return nodeType.Str() == string(conf.PEER_ROLE)
}

// isAlive: checks that the node's configuration has been updated
// since the rquired keep alive time
func (c *CrdtMeshManager) isAlive(nodeId string) bool {
	node, err := c.doc.Path("nodes").Map().Get(nodeId)

	if err != nil || node.Kind() != automerge.KindMap {
		return false
	}

	timestamp, err := node.Map().Get("timestamp")

	if err != nil || timestamp.Kind() != automerge.KindInt64 {
		return false
	}

	keepAliveTime := timestamp.Int64()
	return (time.Now().Unix() - keepAliveTime) < int64(c.conf.DeadTime)
}

func (c *CrdtMeshManager) GetPeers() []string {
	keys, _ := c.doc.Path("nodes").Map().Keys()

	keys = lib.Filter(keys, func(publicKey string) bool {
		return c.isPeer(publicKey) && c.isAlive(publicKey)
	})

	return keys
}

// GetMesh(): Converts the document into a struct
func (c *CrdtMeshManager) GetMesh() (mesh.MeshSnapshot, error) {
	changes, err := c.doc.Changes(c.lastCacheHash)

	if err != nil {
		return nil, err
	}

	if c.cache == nil || len(changes) > 0 {
		c.lastCacheHash = c.LastHash
		cache, err := automerge.As[*MeshCrdt](c.doc.Root())

		if err != nil {
			return nil, err
		}

		c.cache = cache
	}

	return c.cache, nil
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
	manager.cache = nil
	return &manager, nil
}

// NodeExists: returns true if the node exists. Returns false
func (m *CrdtMeshManager) NodeExists(key string) bool {
	node, err := m.doc.Path("nodes").Map().Get(key)
	return node.Kind() == automerge.KindMap && err == nil
}

func (m *CrdtMeshManager) GetNode(endpoint string) (mesh.MeshNode, error) {
	node, err := m.doc.Path("nodes").Map().Get(endpoint)

	if node.Kind() != automerge.KindMap {
		return nil, fmt.Errorf("GetNode: something went wrong %s is not a map type")
	}

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

func (m *CrdtMeshManager) SetAlias(nodeId string, alias string) error {
	node, err := m.doc.Path("nodes").Map().Get(nodeId)

	if err != nil {
		return err
	}

	if node.Kind() != automerge.KindMap {
		return fmt.Errorf("%s does not exist", nodeId)
	}

	err = node.Map().Set("alias", alias)

	if err == nil {
		logging.Log.WriteInfof("Updated Alias for %s to %s", nodeId, alias)
	}

	return err
}

func (m *CrdtMeshManager) AddService(nodeId, key, value string) error {
	node, err := m.doc.Path("nodes").Map().Get(nodeId)

	if err != nil || node.Kind() != automerge.KindMap {
		return fmt.Errorf("AddService: node %s does not exist", nodeId)
	}

	service, err := node.Map().Get("services")

	if err != nil {
		return err
	}

	if service.Kind() != automerge.KindMap {
		return fmt.Errorf("AddService: services property does not exist in node")
	}

	return service.Map().Set(key, value)
}

func (m *CrdtMeshManager) RemoveService(nodeId, key string) error {
	node, err := m.doc.Path("nodes").Map().Get(nodeId)

	if err != nil || node.Kind() != automerge.KindMap {
		return fmt.Errorf("RemoveService: node %s does not exist", nodeId)
	}

	service, err := node.Map().Get("services")

	if err != nil {
		return err
	}

	if service.Kind() != automerge.KindMap {
		return fmt.Errorf("services property does not exist")
	}

	err = service.Map().Delete(key)

	if err != nil {
		return fmt.Errorf("service %s does not exist", key)
	}

	return nil
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

// DeleteRoutes deletes the specified routes
func (m *CrdtMeshManager) RemoveRoutes(nodeId string, routes ...string) error {
	nodeVal, err := m.doc.Path("nodes").Map().Get(nodeId)

	if err != nil {
		return err
	}

	if nodeVal.Kind() != automerge.KindMap {
		return fmt.Errorf("node is not a map")
	}

	routeMap, err := nodeVal.Map().Get("routes")

	if err != nil {
		return err
	}

	for _, route := range routes {
		err = routeMap.Map().Delete(route)
	}

	return err
}

func (m *CrdtMeshManager) GetSyncer() mesh.MeshSyncer {
	return NewAutomergeSync(m)
}

func (m *CrdtMeshManager) Prune(pruneTime int) error {
	nodes, err := m.doc.Path("nodes").Get()

	if err != nil {
		return err
	}

	if nodes.Kind() != automerge.KindMap {
		return errors.New("node must be a map")
	}

	values, err := nodes.Map().Values()

	if err != nil {
		return err
	}

	deletionNodes := make([]string, 0)

	for nodeId, node := range values {
		if node.Kind() != automerge.KindMap {
			return errors.New("node must be a map")
		}

		nodeMap := node.Map()

		timeStamp, err := nodeMap.Get("timestamp")

		if err != nil {
			return err
		}

		if timeStamp.Kind() != automerge.KindInt64 {
			return errors.New("timestamp is not int64")
		}

		timeValue := timeStamp.Int64()
		nowValue := time.Now().Unix()

		if nowValue-timeValue >= int64(pruneTime) {
			deletionNodes = append(deletionNodes, nodeId)
		}
	}

	for _, node := range deletionNodes {
		logging.Log.WriteInfof("Pruning %s", node)
		nodes.Map().Delete(node)
	}

	return nil
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

func (m *MeshNodeCrdt) GetAlias() string {
	return m.Alias
}

func (m *MeshNodeCrdt) GetServices() map[string]string {
	services := make(map[string]string)

	for key, service := range m.Services {
		services[key] = service
	}

	return services
}

// GetType refers to the type of the node. Peer means that the node is globally accessible
// Client means the node is only accessible through another peer
func (n *MeshNodeCrdt) GetType() conf.NodeType {
	return conf.NodeType(n.Type)
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
			Alias:        node.Alias,
			Services:     node.GetServices(),
			Type:         node.Type,
		}
	}

	return nodes
}
