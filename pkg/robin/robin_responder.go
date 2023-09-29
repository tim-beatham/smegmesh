package robin

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"google.golang.org/grpc/peer"
)

type RobinRpc struct {
	rpc.UnimplementedMeshCtrlServerServer
	server *ctrlserver.MeshCtrlServer
}

func nodeToRpcNode(node ctrlserver.MeshNode) *rpc.MeshNode {
	return &rpc.MeshNode{
		PublicKey:  node.PublicKey,
		WgEndpoint: node.WgEndpoint,
		WgHost:     node.WgHost,
		Endpoint:   node.HostEndpoint,
	}
}

func nodesToRpcNodes(nodes map[string]ctrlserver.MeshNode) []*rpc.MeshNode {
	n := len(nodes)
	meshNodes := make([]*rpc.MeshNode, n)

	var i int = 0

	for _, v := range nodes {
		meshNodes[i] = nodeToRpcNode(v)
		i++
	}

	return meshNodes
}

func (m *RobinRpc) GetMesh(ctx context.Context, request *rpc.GetMeshRequest) (*rpc.GetMeshReply, error) {
	mesh, contains := m.server.Meshes[request.MeshId]

	if !contains {
		return nil, errors.New("Element is not in the mesh")
	}

	reply := rpc.GetMeshReply{
		MeshId:   request.MeshId,
		MeshNode: nodesToRpcNodes(mesh.Nodes),
	}

	return &reply, nil
}

func (m *RobinRpc) JoinMesh(ctx context.Context, request *rpc.JoinMeshRequest) (*rpc.JoinMeshReply, error) {
	p, _ := peer.FromContext(ctx)

	hostIp, _, err := net.SplitHostPort(p.Addr.String())

	if err != nil {
		return nil, err
	}

	wgIp := request.WgIp

	if wgIp == "" {
		return nil, errors.New("Haven't provided a valid IP address")
	}

	addHostArgs := ctrlserver.AddHostArgs{
		HostEndpoint: hostIp + ":" + strconv.Itoa(int(request.HostPort)),
		PublicKey:    request.PublicKey,
		MeshId:       request.MeshId,
		WgEndpoint:   hostIp + ":" + strconv.Itoa(int(request.WgPort)),
		WgIp:         wgIp,
	}

	err = m.server.AddHost(addHostArgs)

	if err != nil {
		return nil, err
	}

	return &rpc.JoinMeshReply{Success: true, MeshIp: &wgIp}, nil
}

func NewRobinRpc(ctrlServer *ctrlserver.MeshCtrlServer) *RobinRpc {
	return &RobinRpc{server: ctrlServer}
}
