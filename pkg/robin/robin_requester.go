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

type RobinIpc struct {
	Server      *ctrlserver.MeshCtrlServer
	ipAllocator ip.IPAllocator
}

func (n *RobinIpc) CreateMesh(args *ipc.NewMeshArgs, reply *string) error {
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

	n.Server.MeshManager.AddMeshNode(meshId, meshNode)

	if err != nil {
		return err
	}

	*reply = meshId
	return nil
}

func (n *RobinIpc) ListMeshes(_ string, reply *ipc.ListMeshReply) error {
	meshNames := make([]string, len(n.Server.MeshManager.Meshes))

	i := 0
	for _, mesh := range n.Server.MeshManager.Meshes {
		meshNames[i] = mesh.MeshId
		i++
	}

	*reply = ipc.ListMeshReply{Meshes: meshNames}
	return nil
}

func (n *RobinIpc) JoinMesh(args ipc.JoinMeshArgs, reply *string) error {
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

	n.Server.MeshManager.AddMeshNode(args.MeshId, node)
	*reply = strconv.FormatBool(true)
	return nil
}

func (n *RobinIpc) GetMesh(meshId string, reply *ipc.GetMeshReply) error {
	mesh := n.Server.MeshManager.GetMesh(meshId)
	meshSnapshot, err := mesh.GetCrdt()

	if err != nil {
		return err
	}

	if mesh != nil {
		nodes := make([]ctrlserver.MeshNode, len(meshSnapshot.Nodes))

		i := 0
		for _, node := range meshSnapshot.Nodes {
			node := ctrlserver.MeshNode{
				HostEndpoint: node.HostEndpoint,
				WgEndpoint:   node.WgEndpoint,
				PublicKey:    node.PublicKey,
				WgHost:       node.WgHost,
				Failed:       mesh.HasFailed(node.HostEndpoint),
				Timestamp:    node.Timestamp,
				Routes:       lib.MapKeys(node.Routes),
			}

			nodes[i] = node
			i += 1
		}

		*reply = ipc.GetMeshReply{Nodes: nodes}
	} else {
		return errors.New("mesh does not exist")
	}
	return nil
}

func (n *RobinIpc) EnableInterface(meshId string, reply *string) error {
	err := n.Server.MeshManager.EnableInterface(meshId)

	if err != nil {
		*reply = err.Error()
		return err
	}

	*reply = "up"
	return nil
}

func (n *RobinIpc) GetDOT(meshId string, reply *string) error {
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

func NewRobinIpc(ipcParams RobinIpcParams) RobinIpc {
	return RobinIpc{
		Server:      ipcParams.CtrlServer,
		ipAllocator: ipcParams.Allocator,
	}
}
