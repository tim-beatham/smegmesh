package ctrlserver

import (
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
	Client *wgctrl.Client
	Meshes map[string]Mesh
	IfName string
}
