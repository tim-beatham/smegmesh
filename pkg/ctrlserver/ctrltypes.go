package ctrlserver

import (
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/query"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Represents a WireGuard MeshNode
type MeshNode struct {
	HostEndpoint string
	WgEndpoint   string
	PublicKey    string
	WgHost       string
	Timestamp    int64
	Routes       []string
}

// Represents a WireGuard Mesh
type Mesh struct {
	SharedKey *wgtypes.Key
	Nodes     map[string]MeshNode
}

// Represents a ctrlserver to be used in WireGuard
type MeshCtrlServer struct {
	Client            *wgctrl.Client
	MeshManager       *mesh.MeshManager
	ConnectionManager conn.ConnectionManager
	ConnectionServer  *conn.ConnectionServer
	Conf              *conf.WgMeshConfiguration
	Querier           query.Querier
}
