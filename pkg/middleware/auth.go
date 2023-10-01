package middleware

import (
	"context"
	"errors"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

type AuthRpcProvider struct {
	rpc.UnimplementedAuthenticationServer
	server *ctrlserver.MeshCtrlServer
}

func (a *AuthRpcProvider) JoinMesh(ctx context.Context, in *rpc.JoinAuthMeshRequest) (*rpc.JoinAuthMeshReply, error) {
	meshId := in.MeshId

	if meshId == "" {
		return nil, errors.New("Must specify the meshId")
	}

	token, err := a.server.JwtManager.CreateClaims(in.MeshId, "sharedSecret")

	if err != nil {
		return nil, err
	}

	return &rpc.JoinAuthMeshReply{Success: true, Token: token}, nil
}

func NewAuthProvider(ctrlServer *ctrlserver.MeshCtrlServer) *AuthRpcProvider {
	return &AuthRpcProvider{server: ctrlServer}
}
