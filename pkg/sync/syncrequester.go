package sync

import (
	"context"
	"io"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/conn"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"github.com/tim-beatham/smegmesh/pkg/rpc"
)

// SyncRequester: coordinates the syncing of meshes
type SyncRequester interface {
	SyncMesh(mesh mesh.MeshProvider, meshNode mesh.MeshNode) error
}

type SyncRequesterImpl struct {
	manager           mesh.MeshManager
	connectionManager conn.ConnectionManager
	configuration     *conf.DaemonConfiguration
	errorHdlr         SyncErrorHandler
}

// handleErr: handleGrpc errors
func (s *SyncRequesterImpl) handleErr(mesh mesh.MeshProvider, pubKey string, err error) error {
	ok := s.errorHdlr.Handle(mesh, pubKey, err)

	if ok {
		return nil
	}
	return err
}

// SyncMesh: Proactively send a sync request to the other mesh
func (s *SyncRequesterImpl) SyncMesh(mesh mesh.MeshProvider, meshNode mesh.MeshNode) error {
	endpoint := meshNode.GetHostEndpoint()
	pubKey, _ := meshNode.GetPublicKey()

	peerConnection, err := s.connectionManager.GetConnection(endpoint)

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	c := rpc.NewSyncServiceClient(client)

	syncTimeOut := float64(s.configuration.SyncInterval) * float64(time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(syncTimeOut))
	defer cancel()

	err = s.syncMesh(mesh, ctx, c)

	if err != nil {
		s.handleErr(mesh, pubKey.String(), err)
	}

	logging.Log.WriteInfof("synced with node: %s meshId: %s\n", endpoint, mesh.GetMeshId())
	return err
}

func (s *SyncRequesterImpl) syncMesh(mesh mesh.MeshProvider, ctx context.Context, client rpc.SyncServiceClient) error {
	stream, err := client.SyncMesh(ctx)

	syncer := mesh.GetSyncer()

	if err != nil {
		return err
	}

	for {
		msg, moreMessages := syncer.GenerateMessage()

		err := stream.Send(&rpc.SyncMeshRequest{MeshId: mesh.GetMeshId(), Changes: msg})

		if err != nil {
			return err
		}

		in, err := stream.Recv()

		if err != nil && err != io.EOF {
			logging.Log.WriteInfof("stream recv error: %s\n", err.Error())
			return err
		}

		if err != io.EOF && len(in.Changes) != 0 {
			err = syncer.RecvMessage(in.Changes)
		}

		if err != nil {
			logging.Log.WriteInfof("syncer recv error: %s\n", err.Error())
			return err
		}

		if !moreMessages {
			break
		}
	}

	syncer.Complete()
	stream.CloseSend()
	return nil
}

type NewSyncRequesterParams struct {
	MeshManager       mesh.MeshManager
	ConnectionManager conn.ConnectionManager
	Configuration     *conf.DaemonConfiguration
}

func NewSyncRequester(params NewSyncRequesterParams) SyncRequester {
	errorHdlr := NewSyncErrorHandler(params.MeshManager, params.ConnectionManager)
	return &SyncRequesterImpl{manager: params.MeshManager,
		connectionManager: params.ConnectionManager,
		configuration:     params.Configuration,
		errorHdlr:         errorHdlr,
	}
}
