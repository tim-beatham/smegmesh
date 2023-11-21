// mesh provides implementation agnostic logic for managing
// the mesh
package mesh

import (
	"net"
	"slices"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	// Data Exchanged Between Peers
	PEER conf.NodeType = "peer"
	// Data Exchanged Between Clients
	CLIENT conf.NodeType = "client"
)

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
	GetRoutes() []string
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
	if node1.GetHostEndpoint() != node2.GetHostEndpoint() {
		return false
	}

	node1Pub, _ := node1.GetPublicKey()
	node2Pub, _ := node2.GetPublicKey()

	if node1Pub != node2Pub {
		return false
	}

	if node1.GetWgEndpoint() != node2.GetWgEndpoint() {
		return false
	}

	if node1.GetWgHost() != node2.GetWgHost() {
		return false
	}

	if !slices.Equal(node1.GetRoutes(), node2.GetRoutes()) {
		return false
	}

	if node1.GetIdentifier() != node2.GetIdentifier() {
		return false
	}

	if node1.GetDescription() != node2.GetDescription() {
		return false
	}

	if node1.GetAlias() != node2.GetAlias() {
		return false
	}

	return true
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
	AddRoutes(nodeId string, route ...string) error
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
	GetPeers() []string
}

// HostParameters contains the IDs of a node
type HostParameters struct {
	HostEndpoint string
}

// MeshProviderFactoryParams parameters required to build a mesh provider
type MeshProviderFactoryParams struct {
	DevName string
	MeshId  string
	Port    int
	Conf    *conf.WgMeshConfiguration
	Client  *wgctrl.Client
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
	Role      conf.NodeType
}

// MeshBuilder build the hosts mesh node for it to be added to the mesh
type MeshNodeFactory interface {
	Build(params *MeshNodeFactoryParams) MeshNode
}
