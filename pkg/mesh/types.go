// mesh provides implementation agnostic logic for managing
// the mesh
package mesh

import (
	"net"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	// GetMesh() returns a snapshot of the mesh provided by the mesh provider
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
	// Record that we have changges and save the corresponding changes
	SaveChanges()
	// UpdateTimeStamp: update the timestamp of the given node
	UpdateTimeStamp(nodeId string) error
	// AddRoutes: adds routes to the given node
	AddRoutes(nodeId string, route ...string) error
	GetSyncer() MeshSyncer
}

// HostParameters contains the IDs of a node
type HostParameters struct {
	HostEndpoint string
	// TODO: Contain the WireGuard identifier in this
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
