package sync

import (
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SyncErrorHandler: Handles errors when attempting to sync
type SyncErrorHandler interface {
	Handle(meshId string, endpoint string, err error) bool
}

// SyncErrorHandlerImpl Is an implementation of the SyncErrorHandler
type SyncErrorHandlerImpl struct {
	meshManager *mesh.MeshManager
}

func (s *SyncErrorHandlerImpl) incrementFailedCount(meshId string, endpoint string) bool {
	mesh := s.meshManager.GetMesh(meshId)

	if mesh == nil {
		return false
	}

	return true
}

func (s *SyncErrorHandlerImpl) Handle(meshId string, endpoint string, err error) bool {
	errStatus, _ := status.FromError(err)

	logging.Log.WriteInfof("Handled gRPC error: %s", errStatus.Message())

	switch errStatus.Code() {
	case codes.Unavailable, codes.Unknown, codes.DeadlineExceeded, codes.Internal, codes.NotFound:
		return s.incrementFailedCount(meshId, endpoint)
	}

	return false
}

func NewSyncErrorHandler(m *mesh.MeshManager) SyncErrorHandler {
	return &SyncErrorHandlerImpl{meshManager: m}
}
