package middleware

import (
	"context"
	"errors"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

// AuthRpcProvider implements the AuthRpcProvider service
type AuthRpcProvider struct {
	rpc.UnimplementedAuthenticationServer
}

// JoinMesh handles a JoinMeshRequest. Succeeds by stating the node managed to join the mesh
// or returns an error if it failed
func (a *AuthRpcProvider) JoinMesh(ctx context.Context, in *rpc.JoinAuthMeshRequest) (*rpc.JoinAuthMeshReply, error) {
	meshId := in.MeshId

	if meshId == "" {
		return nil, errors.New("Must specify the meshId")
	}

	logging.Log.WriteInfof("MeshID: " + in.MeshId)

	var token string = ""
	return &rpc.JoinAuthMeshReply{Success: true, Token: &token}, nil
}
