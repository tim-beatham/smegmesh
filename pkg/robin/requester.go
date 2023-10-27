package robin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"github.com/tim-beatham/wgmesh/pkg/wg"
)

type IpcHandler struct {
	Server      *ctrlserver.MeshCtrlServer
	ipAllocator ip.IPAllocator
}

func (n *IpcHandler) CreateMesh(args *ipc.NewMeshArgs, reply *string) error {
	wg.CreateInterface(args.IfName)

	meshId, err := n.Server.MeshManager.CreateMesh(args.IfName, args.WgPort)

	if err != nil {
		return err
	}

	pubKey, err := n.Server.MeshManager.GetPublicKey(meshId)

	if err != nil {
		return err
	}

	nodeIP, err := n.ipAllocator.GetIP(*pubKey, meshId)

	if err != nil {
		return err
	}

	outBoundIp := lib.GetOutboundIP()

	meshNode := crdt.MeshNodeCrdt{
		HostEndpoint: fmt.Sprintf("%s:%s", outBoundIp.String(), n.Server.Conf.GrpcPort),
		PublicKey:    pubKey.String(),
		WgEndpoint:   fmt.Sprintf("%s:%d", outBoundIp.String(), args.WgPort),
		WgHost:       nodeIP.String() + "/128",
		Routes:       map[string]interface{}{},
	}

	err = n.Server.MeshManager.AddMeshNode(meshId, &meshNode)

	if err != nil {
		return err
	}

	*reply = meshId
	return nil
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

	err = n.Server.MeshManager.AddMesh(args.MeshId, args.IfName, args.Port, meshReply.Mesh)

	if err != nil {
		return err
	}

	pubKey, err := n.Server.MeshManager.GetPublicKey(args.MeshId)

	if err != nil {
		return err
	}

	ipAddr, err := n.ipAllocator.GetIP(*pubKey, args.MeshId)

	if err != nil {
		return err
	}

	logging.Log.WriteInfof("WgIP: " + ipAddr.String())

	outBoundIP := lib.GetOutboundIP()

	node := crdt.MeshNodeCrdt{
		HostEndpoint: fmt.Sprintf("%s:%s", outBoundIP.String(), n.Server.Conf.GrpcPort),
		WgEndpoint:   fmt.Sprintf("%s:%d", outBoundIP.String(), args.Port),
		PublicKey:    pubKey.String(),
		WgHost:       ipAddr.String() + "/128",
		Routes:       make(map[string]interface{}),
	}
	err = n.Server.MeshManager.AddMeshNode(args.MeshId, &node)
	
	if err != nil {
		return err
	}
	
	*reply = strconv.FormatBool(true)
	return nil
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
	Allocator  ip.IPAllocator
}

func NewRobinIpc(ipcParams RobinIpcParams) IpcHandler {
	return IpcHandler{
		Server:      ipcParams.CtrlServer,
		ipAllocator: ipcParams.Allocator,
	}
}
