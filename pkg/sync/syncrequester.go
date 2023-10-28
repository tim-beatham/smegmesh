package sync

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

// SyncRequester: coordinates the syncing of meshes
type SyncRequester interface {
	GetMesh(meshId string, ifName string, port int, endPoint string) error
	SyncMesh(meshid string, endPoint string) error
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
		DevName:   ifName,
		WgPort:    port,
		MeshBytes: reply.Mesh,
	})
	return err
}

func (s *SyncRequesterImpl) handleErr(meshId, endpoint string, err error) error {
	ok := s.errorHdlr.Handle(meshId, endpoint, err)

	if ok {
		return nil
	}

	return err
}

// SyncMesh: Proactively send a sync request to the other mesh
func (s *SyncRequesterImpl) SyncMesh(meshId, endpoint string) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = syncMesh(mesh, ctx, c)

	if err != nil {
		return s.handleErr(meshId, endpoint, err)
	}

	logging.Log.WriteInfof("Synced with node: %s meshId: %s\n", endpoint, meshId)
	return nil
}

func syncMesh(mesh mesh.MeshProvider, ctx context.Context, client rpc.SyncServiceClient) error {
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
	errorHdlr := NewSyncErrorHandler(s.MeshManager)
	return &SyncRequesterImpl{server: s, errorHdlr: errorHdlr}
}
