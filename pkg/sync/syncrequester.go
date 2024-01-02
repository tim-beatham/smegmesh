package sync

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"github.com/tim-beatham/smegmesh/pkg/rpc"
)

// SyncRequester: coordinates the syncing of meshes
type SyncRequester interface {
	GetMesh(meshId string, ifName string, port int, endPoint string) error
	SyncMesh(meshid string, meshNode mesh.MeshNode) error
}

type SyncRequesterImpl struct {
	server    *ctrlserver.MeshCtrlServer
	errorHdlr SyncErrorHandler
}

// GetMesh: Retrieves the local state of the mesh at the endpoint
func (s *SyncRequesterImpl) GetMesh(meshId string, ifName string, port int, endPoint string) error {
	peerConnection, err := s.server.ConnectionManager.GetConnection(endPoint)

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	c := rpc.NewSyncServiceClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := c.GetConf(ctx, &rpc.GetConfRequest{MeshId: meshId})

	if err != nil {
		return err
	}

	err = s.server.MeshManager.AddMesh(&mesh.AddMeshParams{
		MeshId:    meshId,
		WgPort:    port,
		MeshBytes: reply.Mesh,
	})
	return err
}

func (s *SyncRequesterImpl) handleErr(meshId, pubKey string, err error) error {
	ok := s.errorHdlr.Handle(meshId, pubKey, err)

	if ok {
		return nil
	}

	return err
}

// SyncMesh: Proactively send a sync request to the other mesh
func (s *SyncRequesterImpl) SyncMesh(meshId string, meshNode mesh.MeshNode) error {
	endpoint := meshNode.GetHostEndpoint()
	pubKey, _ := meshNode.GetPublicKey()

	peerConnection, err := s.server.ConnectionManager.GetConnection(endpoint)

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	mesh := s.server.MeshManager.GetMesh(meshId)

	if mesh == nil {
		return errors.New("mesh does not exist")
	}

	c := rpc.NewSyncServiceClient(client)

	syncTimeOut := float64(s.server.Conf.SyncTime) * float64(time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(syncTimeOut))
	defer cancel()

	err = s.syncMesh(mesh, ctx, c)

	if err != nil {
		s.handleErr(meshId, pubKey.String(), err)
	}

	logging.Log.WriteInfof("Synced with node: %s meshId: %s\n", endpoint, meshId)
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
			logging.Log.WriteInfof("Stream recv error: %s\n", err.Error())
			return err
		}

		if err != io.EOF && len(in.Changes) != 0 {
			err = syncer.RecvMessage(in.Changes)
		}

		if err != nil {
			logging.Log.WriteInfof("Syncer recv error: %s\n", err.Error())
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

func NewSyncRequester(s *ctrlserver.MeshCtrlServer) SyncRequester {
	errorHdlr := NewSyncErrorHandler(s.MeshManager, s.ConnectionManager)
	return &SyncRequesterImpl{server: s, errorHdlr: errorHdlr}
}
