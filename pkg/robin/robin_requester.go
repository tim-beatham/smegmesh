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
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"github.com/tim-beatham/wgmesh/pkg/wg"
)

type RobinIpc struct {
	Server      *ctrlserver.MeshCtrlServer
	ipAllocator ip.IPAllocator
}

func (n *RobinIpc) CreateMesh(name string, reply *string) error {
	wg.CreateInterface("wgmesh")

	meshId, err := n.Server.MeshManager.CreateMesh("wgmesh")

	pubKey, err := n.Server.MeshManager.GetPublicKey(meshId)
	nodeIP, err := n.ipAllocator.GetIP(*pubKey, meshId)

	if err != nil {
		return err
	}

	outBoundIp := lib.GetOutboundIP()

	meshNode := crdt.MeshNodeCrdt{
		HostEndpoint: outBoundIp.String() + ":8080",
		PublicKey:    pubKey.String(),
		WgEndpoint:   outBoundIp.String() + ":51820",
		WgHost:       nodeIP.String() + "/128",
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
		fmt.Println(meshNames[i])
		i++
	}

	*reply = ipc.ListMeshReply{Meshes: meshNames}
	return nil
}

func (n *RobinIpc) Authenticate(meshId, endpoint string) error {
	peerConnection, err := n.Server.ConnectionManager.AddConnection(endpoint)

	if err != nil {
		return err
	}

	err = peerConnection.Authenticate(meshId)

	if err != nil {
		return err
	}

	return err
}

func (n *RobinIpc) updatePeers(meshId string) error {
	theMesh := n.Server.MeshManager.GetMesh(meshId)

	if theMesh == nil {
		return errors.New("the mesh does not exist")
	}

	snapshot, _ := theMesh.GetCrdt()
	publicKey, err := n.Server.MeshManager.GetPublicKey(meshId)

	if err != nil {
		return err
	}

	for nodeKey, node := range snapshot.Nodes {
		logging.InfoLog.Println(nodeKey)
		if nodeKey == publicKey.String() {
			continue
		}

		var reply string
		err := n.JoinMesh(ipc.JoinMeshArgs{MeshId: meshId, IpAdress: node.HostEndpoint}, &reply)

		if err != nil {
			logging.InfoLog.Println(err)
			return err
		}
	}

	return nil
}

func (n *RobinIpc) JoinMesh(args ipc.JoinMeshArgs, reply *string) error {
	err := n.Authenticate(args.MeshId, args.IpAdress)

	if err != nil {
		return err
	}

	peerConnection, err := n.Server.ConnectionManager.GetConnection(args.IpAdress)

	if err != nil {
		return err
	}

	err = peerConnection.Connect()

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	c := rpc.NewMeshCtrlServerClient(client)

	authContext, err := peerConnection.CreateAuthContext(args.MeshId)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(authContext, time.Second)
	defer cancel()

	meshReply, err := c.GetMesh(ctx, &rpc.GetMeshRequest{MeshId: args.MeshId})

	if err != nil {
		return err
	}

	err = n.Server.MeshManager.AddMesh(args.MeshId, "wgmesh", meshReply.Mesh)

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

	logging.InfoLog.Println("WgIP: " + ipAddr.String())

	outBoundIP := lib.GetOutboundIP()

	node := crdt.MeshNodeCrdt{
		HostEndpoint: outBoundIP.String() + ":8080",
		WgEndpoint:   outBoundIP.String() + ":51820",
		PublicKey:    pubKey.String(),
		WgHost:       ipAddr.String() + "/128",
	}

	n.Server.MeshManager.AddMeshNode(args.MeshId, node)
	mesh := n.Server.MeshManager.GetMesh(args.MeshId)

	joinMeshRequest := rpc.JoinMeshRequest{
		MeshId:  args.MeshId,
		Changes: mesh.SaveChanges(),
	}

	joinReply, err := c.JoinMesh(ctx, &joinMeshRequest)

	if err != nil {
		return err
	}

	if joinReply.GetSuccess() {
		err = n.updatePeers(args.MeshId)
	}

	*reply = strconv.FormatBool(joinReply.GetSuccess())
	return nil
}

func (n *RobinIpc) GetMesh(meshId string, reply *ipc.GetMeshReply) error {
	mesh := n.Server.MeshManager.GetMesh(meshId)
	meshSnapshot, err := mesh.GetCrdt()

	if err != nil {
		return err
	}

	if mesh != nil {
		nodes := make([]crdt.MeshNodeCrdt, len(meshSnapshot.Nodes))

		i := 0
		for _, n := range meshSnapshot.Nodes {
			nodes[i] = n
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
	fmt.Println("reached")

	if err != nil {
		*reply = err.Error()
		return err
	}

	*reply = "up"
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