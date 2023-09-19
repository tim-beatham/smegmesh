package main

import (
	"context"
	"errors"
	"fmt"
	"net"

	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/ipc"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/rpc"
	wg "github.com/tim-beatham/wgmesh/pkg/wg"
	"google.golang.org/grpc"
)

type meshCtrlServer struct {
	rpc.UnimplementedMeshCtrlServerServer
	server *ctrlserver.MeshCtrlServer
}

func newServer(ctrl *ctrlserver.MeshCtrlServer) *meshCtrlServer {
	return &meshCtrlServer{server: ctrl}
}

func (m *meshCtrlServer) GetMesh(ctx context.Context, request *rpc.GetMeshRequest) (*rpc.GetMeshReply, error) {
	mesh, contains := m.server.Meshes[request.MeshId]

	if !contains {
		return nil, errors.New("Element is not in the mesh")
	}
	return &rpc.GetMeshReply{MeshId: mesh.SharedKey.String()}, nil
}

func main() {
	wgClient, err := wg.CreateClient("wgmesh")

	if err != nil {
		fmt.Println(err)
		return
	}

	ctrlServer := ctrlserver.NewCtrlServer("0.0.0.0", 21910, wgClient)

	fmt.Println("Running IPC Handler")
	go ipc.RunIpcHandler(ctrlServer)

	fmt.Println("Running gRPC server")

	grpc := grpc.NewServer()

	rpcServer := newServer(ctrlServer)
	rpc.RegisterMeshCtrlServerServer(grpc, rpcServer)

	lis, err := net.Listen("tcp", ":8080")
	if err := grpc.Serve(lis); err != nil {
		fmt.Print(err.Error())
	}
}
