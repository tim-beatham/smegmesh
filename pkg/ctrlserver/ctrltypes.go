package ctrlserver

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

/*
 * Represents a WireGuard node
 */
type MeshNode struct {
	Host     string
	CtrlPort string
	WgPort   string
	WgHost   string
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
	Host   string
	Port   int
	Client *wgctrl.Client
	Meshes map[string]Mesh
}
