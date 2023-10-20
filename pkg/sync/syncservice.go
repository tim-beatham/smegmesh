// sync merges shared state between two nodes
package sync

import (
	"context"
	"errors"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

type SyncServiceImpl struct {
	rpc.UnimplementedSyncServiceServer
	Server *ctrlserver.MeshCtrlServer
}

// GetMesh: Gets a nodes local mesh configuration as a CRDT
func (s *SyncServiceImpl) GetConf(context context.Context, request *rpc.GetConfRequest) (*rpc.GetConfReply, error) {
	mesh := s.Server.MeshManager.GetMesh(request.MeshId)

	if mesh == nil {
		return nil, errors.New("mesh does not exist")
	}

	meshBytes := mesh.Save()

	reply := rpc.GetConfReply{
		Mesh: meshBytes,
	}

	return &reply, nil
}

// Sync: Pings a node and syncs the mesh configuration with the other node
func (s *SyncServiceImpl) SyncMesh(conext context.Context, request *rpc.SyncMeshRequest) (*rpc.SyncMeshReply, error) {
	mesh := s.Server.MeshManager.GetMesh(request.MeshId)

	if mesh == nil {
		return nil, errors.New("mesh does not exist")
	}

	err := s.Server.MeshManager.UpdateMesh(request.MeshId, request.Changes)

	if err != nil {
		return nil, err
	}

	return &rpc.SyncMeshReply{Success: true}, nil
}

func NewSyncService(server *ctrlserver.MeshCtrlServer) *SyncServiceImpl {
	return &SyncServiceImpl{Server: server}
}
