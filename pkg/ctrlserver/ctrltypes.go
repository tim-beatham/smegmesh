package ctrlserver

import (
	"net"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/conn"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"github.com/tim-beatham/smegmesh/pkg/query"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshRoute struct {
	Destination string
	Path        []string
}

// Represents the WireGuard configuration attached to the node
type WireGuardStats struct {
	AllowedIPs                  []string
	TransmitBytes               int64
	ReceivedBytes               int64
	PersistentKeepAliveInterval time.Duration
}

// Represents a WireGuard MeshNode
type MeshNode struct {
	HostEndpoint string
	WgEndpoint   string
	PublicKey    string
	WgHost       string
	Timestamp    int64
	Routes       []MeshRoute
	Description  string
	Alias        string
	Services     map[string]string
	Stats        WireGuardStats
}

// Represents a WireGuard Mesh
type Mesh struct {
	SharedKey *wgtypes.Key
	Nodes     map[string]MeshNode
}

type CtrlServer interface {
	GetConfiguration() *conf.DaemonConfiguration
	GetClient() *wgctrl.Client
	GetQuerier() query.Querier
	GetMeshManager() mesh.MeshManager
	Close() error
	GetConnectionManager() conn.ConnectionManager
}

// Represents a ctrlserver to be used in WireGuard
type MeshCtrlServer struct {
	Client            *wgctrl.Client
	MeshManager       mesh.MeshManager
	ConnectionManager conn.ConnectionManager
	ConnectionServer  *conn.ConnectionServer
	Conf              *conf.DaemonConfiguration
	Querier           query.Querier
}

// NewCtrlNode create an instance of a ctrl node to send over an
// IPC call
func NewCtrlNode(provider mesh.MeshProvider, node mesh.MeshNode) *MeshNode {
	pubKey, _ := node.GetPublicKey()

	ctrlNode := MeshNode{
		HostEndpoint: node.GetHostEndpoint(),
		WgEndpoint:   node.GetWgEndpoint(),
		PublicKey:    pubKey.String(),
		WgHost:       node.GetWgHost().String(),
		Timestamp:    node.GetTimeStamp(),
		Routes: lib.Map(node.GetRoutes(), func(r mesh.Route) MeshRoute {
			return MeshRoute{
				Destination: r.GetDestination().String(),
				Path:        r.GetPath(),
			}
		}),
		Description: node.GetDescription(),
		Alias:       node.GetAlias(),
		Services:    node.GetServices(),
	}

	device, err := provider.GetDevice()

	if err != nil {
		return &ctrlNode
	}

	peers := lib.Filter(device.Peers, func(p wgtypes.Peer) bool {
		return p.PublicKey.String() == pubKey.String()
	})

	if len(peers) > 0 {
		peer := peers[0]

		stats := WireGuardStats{
			AllowedIPs: lib.Map(peer.AllowedIPs, func(i net.IPNet) string {
				return i.String()
			}),
			TransmitBytes:               peer.TransmitBytes,
			ReceivedBytes:               peer.ReceiveBytes,
			PersistentKeepAliveInterval: peer.PersistentKeepaliveInterval,
		}

		ctrlNode.Stats = stats
	}

	return &ctrlNode
}
