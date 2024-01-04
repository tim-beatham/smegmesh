package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tim-beatham/smegmesh/pkg/ipc"
	"github.com/tim-beatham/smegmesh/pkg/what8words"
)

// Route is an advertised route in the data store
type Route struct {
	// Prefix is the advertised route prefix
	Prefix string `json:"prefix"`
	// Path is the hops the destination
	Path []string `json:"path"`
}

// SmegStats is the WireGuard stats that the underlying host
// has sent to the peer
type SmegStats struct {
	// TotalTransmit number of bytes sent to the peer
	TotalTransmit int64 `json:"totalTransmit"`
	// TotalReceived number of bytes received from the peer
	TotalReceived int64 `json:"totalReceived"`
	// KeepAliveInterval WireGuard keepalive interval that is sent to the host
	KeepAliveInterval time.Duration `json:"keepaliveInterval"`
	// AllowsIps is the allowed path to the destination
	AllowedIps []string `json:"allowedIps"`
}

// SmegNode is a node in the mesh network
type SmegNode struct {
	// Alias is the human readable name that the node is assocaited with
	Alias string `json:"alias"`
	// WgHost is the WireGuard IP address of the node. This is an IPv6
	// address
	WgHost string `json:"wgHost"`
	// WgEndpoint is the physical endpoint of the host that packets
	// are forwarded to
	WgEndpoint string `json:"wgEndpoint"`
	// Endpoint is the control plane endpoint of the host which
	// grpc connections are to be sent along
	Endpoint string `json:"endpoint"`
	// Timestamp is the last time the signified it was alive.
	// if the node is the leader this is evert heartBeatInterval
	// otherwise this is the time the node joined the network
	Timestamp int `json:"timestamp"`
	// Description is the human readable description of the node
	Description string `json:"description"`
	// PublicKey is the WireGuard public key of the node
	PublicKey string `json:"publicKey"`
	// Routes is the routes that the node is advertising
	Routes []Route `json:"routes"`
	// Services is information about services that the node offers
	Services map[string]string `json:"services"`
	// Stats is the WireGuard stats of the node (if any)
	Stats SmegStats `json:"stats"`
}

// SmegMesh encapsulates a single mesh in the API
type SmegMesh struct {
	// MeshId is the mesh id of the network
	MeshId string `json:"meshid"`
	// Nodes is the nodes in the network keyed by their public
	// key
	Nodes map[string]SmegNode `json:"nodes"`
}

// CreateMeshRequest encapsulates a request to create a mesh network
type CreateMeshRequest struct {
	// WgPort is the WireGuard to create the mesh in
	WgPort int `json:"port" binding:"omitempty,gte=1024,lt=65535"`
}

// JoinMeshRequests encapsulates a request to create a mesh network
type JoinMeshRequest struct {
	// WgPort is the WireGuard port to run the service on
	WgPort int `json:"port" binding:"omitempty,gte=1024,lt=65535"`
	// Bootstrap is a bootstrap node to use to join the network
	Bootstrap string `json:"bootstrap" binding:"required"`
	// MeshId is the ID of the mesh to join
	MeshId string `json:"meshid" binding:"required"`
}

// ApiServerConf configuration to instantiate the API server
type ApiServerConf struct {
	// WordsFile to use to map IP to words
	WordsFile string
}

// SmegSever is the GIN api server that runs the service
type SmegServer struct {
	// gin router to use
	router *gin.Engine
	// client to invoke operations
	client *ipc.SmegmeshIpc
	// what8words to use to convert IP to an alias
	words *what8words.What8Words
}

// ApiSever absrtacts the API server
type ApiServer interface {
	Run(addr string) error
}
