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
