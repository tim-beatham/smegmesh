// mesh provides implementation agnostic logic for managing
// the mesh
package mesh

import (
	"net"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Route interface {
	// GetDestination: returns the destination of the route
	GetDestination() *net.IPNet
	// GetHopCount: get the total hopcount of the prefix
	GetHopCount() int
	// GetPath: get a list of AS paths to get to the destination
	GetPath() []string
}

type RouteStub struct {
	Destination *net.IPNet
	HopCount    int
	Path        []string
}

func (r *RouteStub) GetDestination() *net.IPNet {
	return r.Destination
}

func (r *RouteStub) GetHopCount() int {
	return r.HopCount
}

func (r *RouteStub) GetPath() []string {
	return r.Path
}

// MeshNode represents an implementation of a node in a mesh
type MeshNode interface {
	// GetHostEndpoint: gets the gRPC endpoint of the node
	GetHostEndpoint() string
	// GetPublicKey: gets the public key of the node
	GetPublicKey() (wgtypes.Key, error)
	// GetWgEndpoint(): get IP and port of the wireguard endpoint
	GetWgEndpoint() string
	// GetWgHost: get the IP address of the WireGuard node
	GetWgHost() *net.IPNet
	// GetTimestamp: get the UNIX time stamp of the ndoe
	GetTimeStamp() int64
	// GetRoutes: returns the routes that the nodes provides
	GetRoutes() []Route
	// GetIdentifier: returns the identifier of the node
	GetIdentifier() string
	// GetDescription: returns the description for this node
	GetDescription() string
	// GetAlias: associates the node with an alias. Potentially used
	// for DNS and so forth.
	GetAlias() string
	// GetServices: returns a list of services offered by the node
	GetServices() map[string]string
	GetType() conf.NodeType
}

// NodeEquals: determines if two mesh nodes are equivalent to one another
func NodeEquals(node1, node2 MeshNode) bool {
	key1, _ := node1.GetPublicKey()
	key2, _ := node2.GetPublicKey()

	return key1.String() == key2.String()
}

func RouteEquals(route1, route2 Route) bool {
	return route1.GetDestination().String() == route2.GetDestination().String() &&
		route1.GetHopCount() == route2.GetHopCount()
}

func NodeID(node MeshNode) string {
	key, _ := node.GetPublicKey()
	return key.String()
}

type MeshSnapshot interface {
	// GetNodes() returns the nodes in the mesh
	GetNodes() map[string]MeshNode
}

// MeshSyncer syncs two meshes
type MeshSyncer interface {
	GenerateMessage() ([]byte, bool)
	RecvMessage(mesg []byte) error
	Complete()
}

// Mesh: Represents an implementation of a mesh
type MeshProvider interface {
	// AddNode() adds a node to the mesh
	AddNode(node MeshNode)
	// GetMesh() returns a snapshot of the mesh provided by the mesh provider.
	GetMesh() (MeshSnapshot, error)
	// GetMeshId() returns the ID of the mesh network
	GetMeshId() string
	// Save() saves the mesh network
	Save() []byte
	// Load() loads a mesh network
	Load([]byte) error
	// GetDevice() get the device corresponding with the mesh
	GetDevice() (*wgtypes.Device, error)
	// HasChanges returns true if we have changes since last time we synced
	HasChanges() bool
	// Record that we have changes and save the corresponding changes
	SaveChanges()
	// UpdateTimeStamp: update the timestamp of the given node
	UpdateTimeStamp(nodeId string) error
	// AddRoutes: adds routes to the given node
	AddRoutes(nodeId string, route ...Route) error
	// DeleteRoutes: deletes the routes from the node
	RemoveRoutes(nodeId string, route ...string) error
	// GetSyncer: returns the automerge syncer for sync
	GetSyncer() MeshSyncer
	// GetNode get a particular not within the mesh
	GetNode(string) (MeshNode, error)
	// NodeExists: returns true if a particular node exists false otherwise
	NodeExists(string) bool
	// SetDescription: sets the description of this automerge data type
	SetDescription(nodeId string, description string) error
	// SetAlias: set the alias of the nodeId
	SetAlias(nodeId string, alias string) error
	// AddService: adds the service to the given node
	AddService(nodeId, key, value string) error
	// RemoveService: removes the service form the node. throws an error if the service does not exist
	RemoveService(nodeId, key string) error
	// Prune: prunes all nodes that have not updated their timestamp in
	// pruneAmount seconds
	Prune(pruneAmount int) error
	// GetPeers: get a list of contactable peers
	GetPeers() []string
	// GetRoutes(): Get all unique routes. Where the route with the least hop count is chosen
	GetRoutes(targetNode string) (map[string]Route, error)
	// RemoveNode(): remove the node from the mesh
	RemoveNode(nodeId string) error
}

// HostParameters contains the IDs of a node
type HostParameters struct {
	PrivateKey *wgtypes.Key
}

// GetPublicKey: gets the public key of the node
func (h *HostParameters) GetPublicKey() string {
	return h.PrivateKey.PublicKey().String()
}

// MeshProviderFactoryParams parameters required to build a mesh provider
type MeshProviderFactoryParams struct {
	DevName string
	MeshId  string
	Port    int
	Conf    *conf.WgMeshConfiguration
	Client  *wgctrl.Client
	NodeID  string
}

// MeshProviderFactory creates an instance of a mesh provider
type MeshProviderFactory interface {
	CreateMesh(params *MeshProviderFactoryParams) (MeshProvider, error)
}

// MeshNodeFactoryParams are the parameters required to construct
// a mesh node
type MeshNodeFactoryParams struct {
	PublicKey *wgtypes.Key
	NodeIP    net.IP
	WgPort    int
	Endpoint  string
}

// MeshBuilder build the hosts mesh node for it to be added to the mesh
type MeshNodeFactory interface {
	Build(params *MeshNodeFactoryParams) MeshNode
}
