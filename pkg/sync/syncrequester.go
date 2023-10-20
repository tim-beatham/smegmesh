package sync

import (
	"context"
	"errors"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

// SyncRequester: coordinates the syncing of meshes
type SyncRequester interface {
	GetMesh(meshId string, endPoint string) error
	SyncMesh(meshid string, endPoint string) error
}

type SyncRequesterImpl struct {
	server *ctrlserver.MeshCtrlServer
}

func (s *SyncRequesterImpl) Authenticate(meshId, endpoint string) error {

	peerConnection, err := s.server.ConnectionManager.AddConnection(endpoint)

	if err != nil {
		return err
	}

	err = peerConnection.Authenticate(meshId)

	if err != nil {
		return err
	}

	return err
}

// GetMesh: Retrieves the local state of the mesh at the endpoint
func (s *SyncRequesterImpl) GetMesh(meshId string, endPoint string) error {
	peerConnection, err := s.server.ConnectionManager.GetConnection(endPoint)

	if err != nil {
		return err
	}

	err = peerConnection.Connect()

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	c := rpc.NewSyncServiceClient(client)
	authContext, err := peerConnection.CreateAuthContext(meshId)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(authContext, time.Second)
	defer cancel()

	reply, err := c.GetConf(ctx, &rpc.GetConfRequest{MeshId: meshId})

	if err != nil {
		return err
	}

	err = s.server.MeshManager.AddMesh(meshId, s.server.Conf.IfName, reply.Mesh)
	return err
}

// SyncMesh: Proactively send a sync request to the other mesh
func (s *SyncRequesterImpl) SyncMesh(meshId, endpoint string) error {
	if !s.server.ConnectionManager.HasConnection(endpoint) {
		s.Authenticate(meshId, endpoint)
	}

	peerConnection, err := s.server.ConnectionManager.GetConnection(endpoint)

	if err != nil {
		return err
	}

	err = peerConnection.Connect()

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	authContext, err := peerConnection.CreateAuthContext(meshId)

	if err != nil {
		return err
	}

	mesh := s.server.MeshManager.GetMesh(meshId)

	if mesh == nil {
		return errors.New("mesh does not exist")
	}

	syncMeshRequest := rpc.SyncMeshRequest{
		MeshId:  meshId,
		Changes: mesh.SaveChanges(),
	}

	c := rpc.NewSyncServiceClient(client)

	ctx, cancel := context.WithTimeout(authContext, time.Second)
	defer cancel()

	_, err = c.SyncMesh(ctx, &syncMeshRequest)

	if err != nil {
		return err
	}

	logging.InfoLog.Printf("Synced with node: %s meshId: %s\n", endpoint, meshId)
	return nil
}

func NewSyncRequester(s *ctrlserver.MeshCtrlServer) SyncRequester {
	return &SyncRequesterImpl{server: s}
}
