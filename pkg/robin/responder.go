package robin

import (
	"context"
	"errors"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

type WgRpc struct {
	rpc.UnimplementedMeshCtrlServerServer
	Server *ctrlserver.MeshCtrlServer
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

func (m *WgRpc) GetMesh(ctx context.Context, request *rpc.GetMeshRequest) (*rpc.GetMeshReply, error) {
	mesh := m.Server.MeshManager.GetMesh(request.MeshId)

	if mesh == nil {
		return nil, errors.New("mesh does not exist")
	}

	meshBytes := mesh.Save()

	reply := rpc.GetMeshReply{
		Mesh: meshBytes,
	}

	return &reply, nil
}

func (m *WgRpc) JoinMesh(ctx context.Context, request *rpc.JoinMeshRequest) (*rpc.JoinMeshReply, error) {
	return &rpc.JoinMeshReply{Success: true}, nil
}
