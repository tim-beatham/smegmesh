package middleware

import (
	"context"
	"errors"

	"github.com/tim-beatham/wgmesh/pkg/auth"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

type AuthRpcProvider struct {
	rpc.UnimplementedAuthenticationServer
	Manager *auth.JwtManager
}

func (a *AuthRpcProvider) JoinMesh(ctx context.Context, in *rpc.JoinAuthMeshRequest) (*rpc.JoinAuthMeshReply, error) {
	meshId := in.MeshId

	if meshId == "" {
		return nil, errors.New("Must specify the meshId")
	}

	logging.InfoLog.Println("MeshID: " + in.MeshId)
	token, err := a.Manager.CreateClaims(in.MeshId, in.Alias)

	if err != nil {
		return nil, err
	}

	return &rpc.JoinAuthMeshReply{Success: true, Token: token}, nil
}
