package ctrlserver

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

/*
 * Create a new control server instance running
 * on the provided port.
 */
func NewCtrlServer(host string, port int, wgClient *wgctrl.Client) *MeshCtrlServer {
	ctrlServer := new(MeshCtrlServer)
	ctrlServer.Port = port
	ctrlServer.Meshes = make(map[string]Mesh)
	ctrlServer.Host = host
	ctrlServer.Client = wgClient
	return ctrlServer
}

/*
 * Given the meshid returns true if the node is in the mesh
 * false otherwise.
 */
func (server *MeshCtrlServer) IsInMesh(meshId string) bool {
	_, inMesh := server.Meshes[meshId]
	return inMesh
}

func (server *MeshCtrlServer) GetEndpoint() string {
	return server.Host + ":" + strconv.Itoa(server.Port)
}

/*
 * Run the gin server instance
 */
func (server *MeshCtrlServer) Run() bool {
	r := gin.Default()
	r.Run(server.GetEndpoint())
	return true
}

func (server *MeshCtrlServer) CreateMesh() (*Mesh, error) {
	key, err := wgtypes.GenerateKey()

	if err != nil {
		return nil, err
	}

	var mesh Mesh = Mesh{
		SharedKey: &key,
		Nodes:     make(map[string]MeshNode),
	}

	server.Meshes[key.String()] = mesh
	return &mesh, nil
}
