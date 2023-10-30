package robin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

type IpcHandler struct {
	Server      *ctrlserver.MeshCtrlServer
	ipAllocator ip.IPAllocator
}

func (n *IpcHandler) CreateMesh(args *ipc.NewMeshArgs, reply *string) error {
	meshId, err := n.Server.MeshManager.CreateMesh(args.IfName, args.WgPort)

	if err != nil {
		return err
	}

	err = n.Server.MeshManager.AddSelf(&mesh.AddSelfParams{
		MeshId:   meshId,
		WgPort:   args.WgPort,
		Endpoint: args.Endpoint,
	})

	*reply = meshId
	return err
}

func (n *IpcHandler) ListMeshes(_ string, reply *ipc.ListMeshReply) error {
	meshNames := make([]string, len(n.Server.MeshManager.Meshes))

	i := 0
	for meshId, _ := range n.Server.MeshManager.Meshes {
		meshNames[i] = meshId
		i++
	}

	*reply = ipc.ListMeshReply{Meshes: meshNames}
	return nil
}

func (n *IpcHandler) JoinMesh(args ipc.JoinMeshArgs, reply *string) error {
	peerConnection, err := n.Server.ConnectionManager.GetConnection(args.IpAdress)

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	c := rpc.NewMeshCtrlServerClient(client)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	meshReply, err := c.GetMesh(ctx, &rpc.GetMeshRequest{MeshId: args.MeshId})

	if err != nil {
		return err
	}

	err = n.Server.MeshManager.AddMesh(&mesh.AddMeshParams{
		MeshId:    args.MeshId,
		DevName:   args.IfName,
		WgPort:    args.Port,
		MeshBytes: meshReply.Mesh,
	})

	if err != nil {
		return err
	}

	err = n.Server.MeshManager.AddSelf(&mesh.AddSelfParams{
		MeshId:   args.MeshId,
		WgPort:   args.Port,
		Endpoint: args.Endpoint,
	})

	if err != nil {
		return err
	}

	*reply = strconv.FormatBool(true)
	return nil
}

// LeaveMesh leaves a mesh network
func (n *IpcHandler) LeaveMesh(meshId string, reply *string) error {
	err := n.Server.MeshManager.LeaveMesh(meshId)

	if err == nil {
		*reply = fmt.Sprintf("Left Mesh %s", meshId)
	}

	return err
}

func (n *IpcHandler) GetMesh(meshId string, reply *ipc.GetMeshReply) error {
	mesh := n.Server.MeshManager.GetMesh(meshId)
	meshSnapshot, err := mesh.GetMesh()

	if err != nil {
		return err
	}

	if mesh == nil {
		return errors.New("mesh does not exist")
	}
	nodes := make([]ctrlserver.MeshNode, len(meshSnapshot.GetNodes()))

	i := 0
	for _, node := range meshSnapshot.GetNodes() {
		pubKey, _ := node.GetPublicKey()

		if err != nil {
			return err
		}

		node := ctrlserver.MeshNode{
			HostEndpoint: node.GetHostEndpoint(),
			WgEndpoint:   node.GetWgEndpoint(),
			PublicKey:    pubKey.String(),
			WgHost:       node.GetWgHost().String(),
			Timestamp:    node.GetTimeStamp(),
			Routes:       node.GetRoutes(),
		}

		nodes[i] = node
		i += 1
	}

	*reply = ipc.GetMeshReply{Nodes: nodes}
	return nil
}

func (n *IpcHandler) EnableInterface(meshId string, reply *string) error {
	err := n.Server.MeshManager.EnableInterface(meshId)

	if err != nil {
		*reply = err.Error()
		return err
	}

	*reply = "up"
	return nil
}

func (n *IpcHandler) GetDOT(meshId string, reply *string) error {
	g := mesh.NewMeshDotConverter(n.Server.MeshManager)

	result, err := g.Generate(meshId)

	if err != nil {
		return err
	}

	*reply = result
	return nil
}

type RobinIpcParams struct {
	CtrlServer *ctrlserver.MeshCtrlServer
}

func NewRobinIpc(ipcParams RobinIpcParams) IpcHandler {
	return IpcHandler{
		Server: ipcParams.CtrlServer,
	}
}
