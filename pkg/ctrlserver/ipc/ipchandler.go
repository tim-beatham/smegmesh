package ipc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/wg"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

type Mesh struct {
	Server *ctrlserver.MeshCtrlServer
}

/*
 * Create a new WireGuard mesh network
 */
func (n Mesh) CreateNewMesh(name *string, reply *string) error {
	wg.CreateInterface("wgmesh")

	mesh, err := n.Server.CreateMesh()

	if err != nil {
		return err
	}

	*reply = mesh.SharedKey.String()
	return nil
}

func (n Mesh) ListMeshes(name *string, reply *map[string]ctrlserver.Mesh) error {
    meshes := n.Server.Meshes
    *reply = meshes
    return nil
}

func RunIpcHandler(server *ctrlserver.MeshCtrlServer) error {
	if err := os.RemoveAll(SockAddr); err != nil {
		return errors.New("Could not find to address")
	}

	newMeshIpc := new(Mesh)
	newMeshIpc.Server = server
	rpc.Register(newMeshIpc)
	rpc.HandleHTTP()

	l, e := net.Listen("unix", SockAddr)
	if e != nil {
		return e
	}

	http.Serve(l, nil)
	return nil
}
