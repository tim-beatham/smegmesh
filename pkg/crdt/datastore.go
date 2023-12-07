package crdt

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Route struct {
	Destination string
	Path        []string
}

// GetDestination implements mesh.Route.
func (r *Route) GetDestination() *net.IPNet {
	_, ipnet, _ := net.ParseCIDR(r.Destination)
	return ipnet
}

// GetHopCount implements mesh.Route.
func (r *Route) GetHopCount() int {
	return len(r.Path)
}

// GetPath implements mesh.Route.
func (r *Route) GetPath() []string {
	return r.Path
}

type MeshNode struct {
	HostEndpoint string
	WgEndpoint   string
	PublicKey    string
	WgHost       string
	Timestamp    int64
	Routes       map[string]Route
	Alias        string
	Description  string
	Services     map[string]string
	Type         string
	Tombstone    bool
}

// Mark: marks the node is unreachable. This is not broadcast on
// syncrhonisation
func (m *TwoPhaseStoreMeshManager) Mark(nodeId string) {
	m.store.Mark(nodeId)
}

// GetHostEndpoint: gets the gRPC endpoint of the node
func (n *MeshNode) GetHostEndpoint() string {
	return n.HostEndpoint
}

// GetPublicKey: gets the public key of the node
func (n *MeshNode) GetPublicKey() (wgtypes.Key, error) {
	return wgtypes.ParseKey(n.PublicKey)
}

// GetWgEndpoint(): get IP and port of the wireguard endpoint
func (n *MeshNode) GetWgEndpoint() string {
	return n.WgEndpoint
}

// GetWgHost: get the IP address of the WireGuard node
func (n *MeshNode) GetWgHost() *net.IPNet {
	_, ipnet, _ := net.ParseCIDR(n.WgHost)
	return ipnet
}

// GetTimestamp: get the UNIX time stamp of the ndoe
func (n *MeshNode) GetTimeStamp() int64 {
	return n.Timestamp
}

// GetRoutes: returns the routes that the nodes provides
func (n *MeshNode) GetRoutes() []mesh.Route {
	routes := make([]mesh.Route, len(n.Routes))

	for index, route := range lib.MapValues(n.Routes) {
		routes[index] = &Route{
			Destination: route.Destination,
			Path:        route.Path,
		}
	}

	return routes
}

// GetIdentifier: returns the identifier of the node
func (m *MeshNode) GetIdentifier() string {
	ipv6 := m.WgHost[:len(m.WgHost)-4]

	constituents := strings.Split(ipv6, ":")
	constituents = constituents[4:]
	return strings.Join(constituents, ":")
}

// GetDescription: returns the description for this node
func (n *MeshNode) GetDescription() string {
	return n.Description
}

// GetAlias: associates the node with an alias. Potentially used
// for DNS and so forth.
func (n *MeshNode) GetAlias() string {
	return n.Alias
}

// GetServices: returns a list of services offered by the node
func (n *MeshNode) GetServices() map[string]string {
	return n.Services
}

func (n *MeshNode) GetType() conf.NodeType {
	return conf.NodeType(n.Type)
}

type MeshSnapshot struct {
	Nodes map[string]MeshNode
}

// GetNodes() returns the nodes in the mesh
func (m *MeshSnapshot) GetNodes() map[string]mesh.MeshNode {
	newMap := make(map[string]mesh.MeshNode)

	for key, value := range m.Nodes {
		newMap[key] = &MeshNode{
			HostEndpoint: value.HostEndpoint,
			PublicKey:    value.PublicKey,
			WgHost:       value.WgHost,
			WgEndpoint:   value.WgEndpoint,
			Timestamp:    value.Timestamp,
			Routes:       value.Routes,
			Alias:        value.Alias,
			Description:  value.Description,
			Services:     value.Services,
			Type:         value.Type,
		}
	}

	return newMap
}

type TwoPhaseStoreMeshManager struct {
	MeshId    string
	IfName    string
	Client    *wgctrl.Client
	LastClock uint64
	conf      *conf.WgMeshConfiguration
	store     *TwoPhaseMap[string, MeshNode]
}

// AddNode() adds a node to the mesh
func (m *TwoPhaseStoreMeshManager) AddNode(node mesh.MeshNode) {
	crdt, ok := node.(*MeshNode)

	if !ok {
		panic("node must be of type mesh node")
	}

	crdt.Routes = make(map[string]Route)
	crdt.Services = make(map[string]string)
	crdt.Timestamp = time.Now().Unix()

	m.store.Put(crdt.PublicKey, *crdt)
}

// GetMesh() returns a snapshot of the mesh provided by the mesh provider.
func (m *TwoPhaseStoreMeshManager) GetMesh() (mesh.MeshSnapshot, error) {
	return &MeshSnapshot{
		Nodes: m.store.AsMap(),
	}, nil
}

// GetMeshId() returns the ID of the mesh network
func (m *TwoPhaseStoreMeshManager) GetMeshId() string {
	return m.MeshId
}

// Save() saves the mesh network
func (m *TwoPhaseStoreMeshManager) Save() []byte {
	snapshot := m.store.Snapshot()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(*snapshot)

	if err != nil {
		logging.Log.WriteInfof(err.Error())
	}

	return buf.Bytes()
}

// Load() loads a mesh network
func (m *TwoPhaseStoreMeshManager) Load(bs []byte) error {
	buf := bytes.NewBuffer(bs)
	dec := gob.NewDecoder(buf)

	var snapshot TwoPhaseMapSnapshot[string, MeshNode]
	err := dec.Decode(&snapshot)

	m.store.Merge(snapshot)
	return err
}

// GetDevice() get the device corresponding with the mesh
func (m *TwoPhaseStoreMeshManager) GetDevice() (*wgtypes.Device, error) {
	dev, err := m.Client.Device(m.IfName)

	if err != nil {
		return nil, err
	}

	return dev, nil
}

// HasChanges returns true if we have changes since last time we synced
func (m *TwoPhaseStoreMeshManager) HasChanges() bool {
	clockValue := m.store.GetHash()
	return clockValue != m.LastClock
}

// Record that we have changes and save the corresponding changes
func (m *TwoPhaseStoreMeshManager) SaveChanges() {
	clockValue := m.store.GetHash()
	m.LastClock = clockValue
}

// UpdateTimeStamp: update the timestamp of the given node
func (m *TwoPhaseStoreMeshManager) UpdateTimeStamp(nodeId string) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	// Sort nodes by their public key
	peers := m.GetPeers()
	slices.Sort(peers)

	if len(peers) == 0 {
		return nil
	}

	peerToUpdate := peers[0]

	if uint64(time.Now().Unix())-m.store.Clock.GetTimestamp(peerToUpdate) > 3*uint64(m.conf.KeepAliveTime) {
		m.store.Mark(peerToUpdate)

		if len(peers) < 2 {
			return nil
		}

		peerToUpdate = peers[1]
	}

	if peerToUpdate != nodeId {
		return nil
	}

	// Refresh causing node to update it's time stamp
	node := m.store.Get(nodeId)
	node.Timestamp = time.Now().Unix()
	m.store.Put(nodeId, node)
	return nil
}

// AddRoutes: adds routes to the given node
func (m *TwoPhaseStoreMeshManager) AddRoutes(nodeId string, routes ...mesh.Route) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	if len(routes) == 0 {
		return nil
	}

	node := m.store.Get(nodeId)

	changes := false

	for _, route := range routes {
		prevRoute, ok := node.Routes[route.GetDestination().String()]

		if !ok || route.GetHopCount() < prevRoute.GetHopCount() {
			changes = true

			node.Routes[route.GetDestination().String()] = Route{
				Destination: route.GetDestination().String(),
				Path:        route.GetPath(),
			}
		}
	}

	if changes {
		m.store.Put(nodeId, node)
	}

	return nil
}

// DeleteRoutes: deletes the routes from the node
func (m *TwoPhaseStoreMeshManager) RemoveRoutes(nodeId string, routes ...string) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	if len(routes) == 0 {
		return nil
	}

	node := m.store.Get(nodeId)

	for _, route := range routes {
		delete(node.Routes, route)
	}

	return nil
}

// GetSyncer: returns the automerge syncer for sync
func (m *TwoPhaseStoreMeshManager) GetSyncer() mesh.MeshSyncer {
	return NewTwoPhaseSyncer(m)
}

// GetNode get a particular not within the mesh
func (m *TwoPhaseStoreMeshManager) GetNode(nodeId string) (mesh.MeshNode, error) {
	if !m.store.Contains(nodeId) {
		return nil, fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	node := m.store.Get(nodeId)
	return &node, nil
}

// NodeExists: returns true if a particular node exists false otherwise
func (m *TwoPhaseStoreMeshManager) NodeExists(nodeId string) bool {
	return m.store.Contains(nodeId)
}

// SetDescription: sets the description of this automerge data type
func (m *TwoPhaseStoreMeshManager) SetDescription(nodeId string, description string) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	node := m.store.Get(nodeId)
	node.Description = description

	m.store.Put(nodeId, node)
	return nil
}

// SetAlias: set the alias of the nodeId
func (m *TwoPhaseStoreMeshManager) SetAlias(nodeId string, alias string) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	node := m.store.Get(nodeId)
	node.Description = alias

	m.store.Put(nodeId, node)
	return nil
}

// AddService: adds the service to the given node
func (m *TwoPhaseStoreMeshManager) AddService(nodeId string, key string, value string) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	node := m.store.Get(nodeId)
	node.Services[key] = value
	m.store.Put(nodeId, node)
	return nil
}

// RemoveService: removes the service form the node. throws an error if the service does not exist
func (m *TwoPhaseStoreMeshManager) RemoveService(nodeId string, key string) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	node := m.store.Get(nodeId)
	delete(node.Services, key)
	m.store.Put(nodeId, node)
	return nil
}

// Prune: prunes all nodes that have not updated their timestamp in
func (m *TwoPhaseStoreMeshManager) Prune() error {
	m.store.Prune()
	return nil
}

// GetPeers: get a list of contactable peers
func (m *TwoPhaseStoreMeshManager) GetPeers() []string {
	nodes := lib.MapValues(m.store.AsMap())
	nodes = lib.Filter(nodes, func(mn MeshNode) bool {
		if mn.Type != string(conf.PEER_ROLE) {
			return false
		}

		// If the node is marked as unreachable don't consider it a peer.
		// this help to optimize convergence time for unreachable nodes.
		// However advertising it to other nodes could result in flapping.
		if m.store.IsMarked(mn.PublicKey) {
			return false
		}

		return true
	})

	return lib.Map(nodes, func(mn MeshNode) string {
		return mn.PublicKey
	})
}

func (m *TwoPhaseStoreMeshManager) getRoutes(targetNode string) (map[string]Route, error) {
	if !m.store.Contains(targetNode) {
		return nil, fmt.Errorf("getRoute: cannot get route %s does not exist", targetNode)
	}

	node := m.store.Get(targetNode)
	return node.Routes, nil
}

// GetRoutes(): Get all unique routes. Where the route with the least hop count is chosen
func (m *TwoPhaseStoreMeshManager) GetRoutes(targetNode string) (map[string]mesh.Route, error) {
	node, err := m.GetNode(targetNode)

	if err != nil {
		return nil, err
	}

	routes := make(map[string]mesh.Route)

	// Add routes that the node directly has
	for _, route := range node.GetRoutes() {
		routes[route.GetDestination().String()] = route
	}

	// Work out the other routes in the mesh
	for _, node := range m.GetPeers() {
		nodeRoutes, err := m.getRoutes(node)

		if err != nil {
			return nil, err
		}

		for _, route := range nodeRoutes {
			otherRoute, ok := routes[route.GetDestination().String()]

			hopCount := route.GetHopCount()

			if node != targetNode {
				hopCount += 1
			}

			if !ok || route.GetHopCount()+1 < otherRoute.GetHopCount() {
				routes[route.GetDestination().String()] = &Route{
					Destination: route.GetDestination().String(),
					Path:        append(route.GetPath(), m.GetMeshId()),
				}
			}
		}
	}

	return routes, nil
}

// RemoveNode(): remove the node from the mesh
func (m *TwoPhaseStoreMeshManager) RemoveNode(nodeId string) error {
	if !m.store.Contains(nodeId) {
		return fmt.Errorf("datastore: %s does not exist in the mesh", nodeId)
	}

	m.store.Remove(nodeId)
	return nil
}
