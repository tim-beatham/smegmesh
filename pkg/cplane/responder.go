package robin

import (
	"context"
	"errors"

	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
	"github.com/tim-beatham/smegmesh/pkg/rpc"
)

// WgRpc: represents a WireGuard rpc call
type WgRpc struct {
	rpc.UnimplementedMeshCtrlServerServer
	Server *ctrlserver.MeshCtrlServer
}

// GetMesh: serialise the mesh network into bytes
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
