package ipc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	ipcRpc "net/rpc"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/rpc"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	ipctypes "github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

func updateMesh(n *Mesh, meshId string, endPoint string) error {
	conn, err := grpc.Dial(endPoint, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	defer conn.Close()
	c := rpc.NewMeshCtrlServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	getMeshReq := rpc.GetMeshRequest{
		MeshId: meshId,
	}

	r, err := c.GetMesh(ctx, &getMeshReq)

	if err != nil {
		return err
	}

	key, err := wgtypes.ParseKey(meshId)
	if err != nil {
		return err
	}

	mesh := new(ctrlserver.Mesh)
	mesh.Nodes = make(map[string]ctrlserver.MeshNode)
	mesh.SharedKey = &key
	n.Server.Meshes[meshId] = *mesh

	for _, node := range r.GetMeshNode() {
		meshNode := ctrlserver.MeshNode{
			PublicKey:    node.PublicKey,
			HostEndpoint: node.Endpoint,
			WgEndpoint:   node.WgEndpoint,
			WgHost:       node.WgHost,
		}

		n.Server.Meshes[meshId].Nodes[meshNode.HostEndpoint] = meshNode
		ctrlserver.AddWgPeer(n.Server.IfName, n.Server.Client, meshNode)
	}

	return nil
}

func updatePeer(n *Mesh, node ctrlserver.MeshNode, wgHost string, meshId string) error {
	conn, err := grpc.Dial(node.HostEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	defer conn.Close()

	c := rpc.NewMeshCtrlServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dev := n.Server.GetDevice()
	joinMeshReq := rpc.JoinMeshRequest{
		MeshId:    meshId,
		HostPort:  8080,
		PublicKey: dev.PublicKey.String(),
		WgPort:    int32(dev.ListenPort),
		WgIp:      wgHost,
	}

	r, err := c.JoinMesh(ctx, &joinMeshReq)

	if err != nil {
		return err
	}

	if !r.GetSuccess() {
		return errors.New("Could not join the mesh")
	}

	return nil
}

func updatePeers(n *Mesh, meshId string, wgHost string, nodesToExclude []string) error {
	for _, node := range n.Server.Meshes[meshId].Nodes {
		nodeEndpoint := node.HostEndpoint

		if !slices.Contains(nodesToExclude, nodeEndpoint) {
			// Best effort service
			err := updatePeer(n, node, wgHost, meshId)

			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

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

	dev := n.Server.GetDevice()

	fmt.Print("Pub Key:" + dev.PublicKey.String())

	joinMeshReq := rpc.JoinMeshRequest{
		MeshId:    args.MeshId,
		HostPort:  8080,
		PublicKey: dev.PublicKey.String(),
		WgPort:    int32(dev.ListenPort),
	}

	r, err := c.JoinMesh(ctx, &joinMeshReq)

	if err != nil {
		return err
	}

	if r.GetSuccess() {
		updateMesh(&n, args.MeshId, args.IpAdress+":8080")
	}

	err = updatePeers(&n, args.MeshId, r.GetMeshIp(), make([]string, 0))

	*reply = strconv.FormatBool(r.GetSuccess())
	return nil
}

func (n Mesh) GetMesh(meshId string, reply *ipc.GetMeshReply) error {
	mesh, contains := n.Server.Meshes[meshId]

	if contains {
		nodes := make([]ctrlserver.MeshNode, len(mesh.Nodes))

		i := 0
		for _, n := range mesh.Nodes {
			fmt.Println(n.PublicKey)
			nodes[i] = n
			i += 1
		}

		*reply = ipc.GetMeshReply{Nodes: nodes}
	} else {
		return errors.New("mesh does not exist")
	}
	return nil
}

func (n Mesh) EnableInterface(meshId string, reply *string) error {
	err := n.Server.EnableInterface(meshId)

	if err != nil {
		return err
	}

	*reply = "up"
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
