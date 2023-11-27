package ctrlserver

import (
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/query"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type MeshRoute struct {
	Destination string
	Path        []string
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
}

// Represents a WireGuard Mesh
type Mesh struct {
	SharedKey *wgtypes.Key
	Nodes     map[string]MeshNode
}

type CtrlServer interface {
	GetConfiguration() *conf.WgMeshConfiguration
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
	Conf              *conf.WgMeshConfiguration
	Querier           query.Querier
}
