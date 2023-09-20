package rpc

import (
	context "context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type meshCtrlServer struct {
	UnimplementedMeshCtrlServerServer
	server *ctrlserver.MeshCtrlServer
}

func nodeToRpcNode(node ctrlserver.MeshNode) *MeshNode {
	return &MeshNode{
		PublicKey:  node.PublicKey,
		WgEndpoint: node.WgEndpoint,
		WgHost:     node.WgHost,
		Endpoint:   node.HostEndpoint,
	}
}

func nodesToRpcNodes(nodes map[string]ctrlserver.MeshNode) []*MeshNode {
	n := len(nodes)
	meshNodes := make([]*MeshNode, n)

	var i int = 0

	for _, v := range nodes {
		meshNodes[i] = nodeToRpcNode(v)
		i++
	}

	return meshNodes
}

func (m *meshCtrlServer) GetMesh(ctx context.Context, request *GetMeshRequest) (*GetMeshReply, error) {
	mesh, contains := m.server.Meshes[request.MeshId]

	if !contains {
		return nil, errors.New("Element is not in the mesh")
	}

	reply := GetMeshReply{
		MeshId:   request.MeshId,
		MeshNode: nodesToRpcNodes(mesh.Nodes),
	}

	return &reply, nil
}

func (m *meshCtrlServer) JoinMesh(ctx context.Context, request *JoinMeshRequest) (*JoinMeshReply, error) {
	p, _ := peer.FromContext(ctx)
	fmt.Println(p.Addr.String())

	hostIp, _, err := net.SplitHostPort(p.Addr.String())

	if err != nil {
		return nil, err
	}

	addHostArgs := ctrlserver.AddHostArgs{
		HostEndpoint: "[" + hostIp + "]" + ":" + strconv.Itoa(int(request.HostPort)),
		PublicKey:    request.PublicKey,
		MeshId:       request.MeshId,
		WgEndpoint:   "[" + hostIp + "]" + ":" + strconv.Itoa(int(request.WgPort)),
	}

	err = m.server.AddHost(addHostArgs)

	if err != nil {
		return &JoinMeshReply{Success: false}, nil
	}

	fmt.Println("success!")
	return &JoinMeshReply{Success: true}, nil
}

func NewRpcServer(ctlServer *ctrlserver.MeshCtrlServer) *grpc.Server {
	server := &meshCtrlServer{server: ctlServer}
	grpc := grpc.NewServer()
	RegisterMeshCtrlServerServer(grpc, server)
	return grpc
}
