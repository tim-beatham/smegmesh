package ctrlserver

import (
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

/*
 * Represents a WireGuard node
 */
type MeshNode struct {
	HostEndpoint string
	WgEndpoint   string
	PublicKey    string
	WgHost       string
	Failed       bool
}

type Mesh struct {
	SharedKey *wgtypes.Key
	Nodes     map[string]MeshNode
}

/*
 * Defines the mesh control server this node
 * is running
 */
type MeshCtrlServer struct {
	Client            *wgctrl.Client
	MeshManager       *mesh.MeshManger
	ConnectionManager conn.ConnectionManager
	ConnectionServer  *conn.ConnectionServer
	Conf              *conf.WgMeshConfiguration
}
