package rpc

import grpc "google.golang.org/grpc"

func NewRpcServer(rpcServer *grpc.Server, server MeshCtrlServerServer, auth AuthenticationServer) *grpc.Server {
	RegisterMeshCtrlServerServer(rpcServer, server)
	RegisterAuthenticationServer(rpcServer, auth)
	return rpcServer
}
