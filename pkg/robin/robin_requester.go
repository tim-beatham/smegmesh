package robin

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"github.com/tim-beatham/wgmesh/pkg/slaac"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type RobinIpc struct {
	Server *ctrlserver.MeshCtrlServer
}

const MeshIfName = "wgmesh"

func (n *RobinIpc) CreateMesh(name string, reply *string) error {
	wg.CreateInterface(MeshIfName)

	mesh, err := n.Server.CreateMesh()
	ula, _ := slaac.NewULA(n.Server.GetDevice().PublicKey, "0")

	outBoundIp := lib.GetOutboundIP().String()

	addHostArgs := ctrlserver.AddHostArgs{
		HostEndpoint: outBoundIp + ":8080",
		PublicKey:    n.Server.GetDevice().PublicKey.String(),
		WgEndpoint:   outBoundIp + ":51820",
		WgIp:         ula.CGA.GetIpv6().String() + "/128",
		MeshId:       mesh.SharedKey.String(),
	}

	n.Server.AddHost(addHostArgs)

	if err != nil {
		return err
	}

	*reply = mesh.SharedKey.String()
	return nil
}

func (n *RobinIpc) ListMeshes(name string, reply *map[string]ctrlserver.Mesh) error {
	*reply = n.Server.Meshes
	return nil
}

func updateMesh(n *RobinIpc, meshId string, endPoint string) error {
	conn, err := n.Server.Conn.Connect(endPoint)

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
		n.Server.AddWgPeer(meshNode)
	}

	return nil
}

func updatePeer(n *RobinIpc, node ctrlserver.MeshNode, wgHost string, meshId string) error {
	conn, err := n.Server.Conn.Connect(node.HostEndpoint)

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

func (n *RobinIpc) JoinMesh(args ipc.JoinMeshArgs, reply *string) error {
	conn, err := n.Server.Conn.Connect(args.IpAdress + ":8080")

	if err != nil {
		return err
	}

	defer conn.Close()

	c := rpc.NewMeshCtrlServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dev := n.Server.GetDevice()
	ula, _ := slaac.NewULA(dev.PublicKey, "0")

	logging.InfoLog.Println("WgIP: " + ula.CGA.GetIpv6().String())

	joinMeshReq := rpc.JoinMeshRequest{
		MeshId:    args.MeshId,
		HostPort:  8080,
		PublicKey: dev.PublicKey.String(),
		WgPort:    int32(dev.ListenPort),
		WgIp:      ula.CGA.GetIpv6().String() + "/128",
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

func NewRobinIpc(ctrlServer *ctrlserver.MeshCtrlServer) *RobinIpc {
	return &RobinIpc{Server: ctrlServer}
}
