package robin

import (
	"context"
	"errors"
	"slices"
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
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

	outBoundIp := lib.GetOutboundIP().String()

	meshNode := crdt.MeshNodeCrdt{
		HostEndpoint: outBoundIp + ":8080",
		PublicKey:    pubKey.String(),
		WgEndpoint:   outBoundIp + ":51820",
		WgHost:       nodeIP.String() + "/128",
	}

	n.Server.MeshManager.AddMeshNode(meshId, meshNode)

	if err != nil {
		return err
	}

	*reply = meshId
	return nil
}

func (n *RobinIpc) ListMeshes(name string, reply *map[string]ctrlserver.Mesh) error {
	// *reply = n.Server.Meshes
	return nil
}

func updateMesh(n *RobinIpc, meshId string, endPoint string) error {
	peerConn, err := n.Server.ConnectionManager.GetConnection(endPoint)

	if err != nil {
		return err
	}

	conn, err := peerConn.GetClient()

	c := rpc.NewMeshCtrlServerClient(conn)

	authContext, err := peerConn.CreateAuthContext(meshId)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(authContext, time.Second)
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

	err := n.Server.MeshManager.AddMesh(meshId, "wgmesh", r.Mesh)

	if err != nil {
		return err
	}

	return nil
}

func updatePeer(n *RobinIpc, node ctrlserver.MeshNode, wgHost string, meshId string) error {
	err := n.Authenticate(meshId, node.HostEndpoint)

	if err != nil {
		return err
	}

	peerConnection, err := n.Server.ConnectionManager.GetConnection(node.HostEndpoint)

	if err != nil {
		return err
	}

	conn, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	c := rpc.NewMeshCtrlServerClient(conn)

	authContext, err := peerConnection.CreateAuthContext(meshId)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(authContext, time.Second)
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

func updatePeers(n *RobinIpc, meshId string, wgHost string, nodesToExclude []string) error {
	for _, node := range n.Server.Meshes[meshId].Nodes {
		nodeEndpoint := node.HostEndpoint

		if !slices.Contains(nodesToExclude, nodeEndpoint) {
			// Best effort service
			err := updatePeer(n, node, wgHost, meshId)

			if err != nil {
				return err
			}
		}
	}

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

func (n *RobinIpc) JoinMesh(args ipc.JoinMeshArgs, reply *string) error {
	err := n.Authenticate(args.MeshId, args.IpAdress+":8080")

	if err != nil {
		return err
	}

	peerConnection, err := n.Server.ConnectionManager.GetConnection(args.IpAdress + ":8080")

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

	dev := n.Server.GetDevice()

	ipAddr, err := n.ipAllocator.GetIP(n.Server.GetPublicKey(), args.MeshId)

	if err != nil {
		return err
	}

	logging.InfoLog.Println("WgIP: " + ipAddr.String())

	joinMeshReq := rpc.JoinMeshRequest{
		MeshId:    args.MeshId,
		HostPort:  8080,
		PublicKey: dev.PublicKey.String(),
		WgPort:    int32(dev.ListenPort),
		WgIp:      ipAddr.String() + "/128",
	}

	r, err := c.JoinMesh(ctx, &joinMeshReq)

	if err != nil {
		return err
	}

	if r.GetSuccess() {
		updateMesh(n, args.MeshId, args.IpAdress+":8080")
	}

	err = updatePeers(n, args.MeshId, r.GetMeshIp(), make([]string, 0))

	*reply = strconv.FormatBool(r.GetSuccess())
	return nil
}

func (n *RobinIpc) GetMesh(meshId string, reply *ipc.GetMeshReply) error {
	mesh, contains := n.Server.Meshes[meshId]

	if contains {
		nodes := make([]ctrlserver.MeshNode, len(mesh.Nodes))

		i := 0
		for _, n := range mesh.Nodes {
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
	err := n.Server.EnableInterface(meshId)

	if err != nil {
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
