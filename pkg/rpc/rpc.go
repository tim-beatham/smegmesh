package rpc

import grpc "google.golang.org/grpc"

func NewRpcServer(server MeshCtrlServerServer) *grpc.Server {
	grpc := grpc.NewServer()
	RegisterMeshCtrlServerServer(grpc, server)
	return grpc
}
