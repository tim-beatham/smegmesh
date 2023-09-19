package ipc

import (
	"context"
	"errors"
	"net"
	"net/http"
	ipcRpc "net/rpc"
	"os"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/rpc"
	ipctypes "github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (n Mesh) JoinMesh(args *ipctypes.JoinMeshArgs, reply *string) error {
	conn, err := grpc.Dial(args.IpAdress+":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	defer conn.Close()

	c := rpc.NewMeshCtrlServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.GetMesh(ctx, &rpc.GetMeshRequest{MeshId: args.MeshId})

	if err != nil {
		return err
	}

	*reply = r.GetMeshId()
	return nil
}

func RunIpcHandler(server *ctrlserver.MeshCtrlServer) error {
	if err := os.RemoveAll(SockAddr); err != nil {
		return errors.New("Could not find to address")
	}

	newMeshIpc := new(Mesh)
	newMeshIpc.Server = server
	ipcRpc.Register(newMeshIpc)
	ipcRpc.HandleHTTP()

	l, e := net.Listen("unix", SockAddr)
	if e != nil {
		return e
	}

	http.Serve(l, nil)
	return nil
}
