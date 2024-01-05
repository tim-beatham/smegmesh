// sync merges shared state between two nodes
package sync

import (
	"context"
	"errors"
	"io"

	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"github.com/tim-beatham/smegmesh/pkg/rpc"
)

type SyncServiceImpl struct {
	rpc.UnimplementedSyncServiceServer
	MeshManager mesh.MeshManager
}

// GetMesh: Gets a nodes local mesh configuration as a CRDT
func (s *SyncServiceImpl) GetConf(context context.Context, request *rpc.GetConfRequest) (*rpc.GetConfReply, error) {
	mesh := s.MeshManager.GetMesh(request.MeshId)

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
// SyncMesh: syncs the two streams changes
func (s *SyncServiceImpl) SyncMesh(stream rpc.SyncService_SyncMeshServer) error {
	var meshId = ""
	var syncer mesh.MeshSyncer = nil

	for {
		in, err := stream.Recv()

		if err == io.EOF {
			if syncer != nil {
				syncer.Complete()
			}
			return nil
		}

		if err != nil {
			return err
		}

		if len(meshId) == 0 {
			meshId = in.MeshId

			mesh := s.MeshManager.GetMesh(meshId)

			if mesh == nil {
				return errors.New("mesh does not exist")
			}

			syncer = mesh.GetSyncer()
		} else if meshId != in.MeshId {
			return errors.New("differing meshids")
		}

		if syncer == nil {
			return errors.New("syncer should not be nil")
		}

		msg, moreMessages := syncer.GenerateMessage()

		if err = stream.Send(&rpc.SyncMeshReply{Success: true, Changes: msg}); err != nil {
			return err
		}

		if len(in.Changes) != 0 {
			if err = syncer.RecvMessage(in.Changes); err != nil {
				return err
			}
		}

		if !moreMessages || err == io.EOF {
			if syncer != nil {
				syncer.Complete()
			}

			return nil
		}
	}
}
