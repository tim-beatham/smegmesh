/*
 * RPC component of the server
 */
package ctrlserver

import (
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"google.golang.org/grpc"
)

func NewRpcServer(server rpc.MeshCtrlServerServer) *grpc.Server {
	grpc := grpc.NewServer()
	rpc.RegisterMeshCtrlServerServer(grpc, server)
	return grpc
}
