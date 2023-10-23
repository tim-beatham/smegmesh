package sync

import (
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SyncErrorHandler interface {
	Handle(meshId string, endpoint string, err error) bool
}

type SyncErrorHandlerImpl struct {
	meshManager *mesh.MeshManger
}

func (s *SyncErrorHandlerImpl) incrementFailedCount(meshId string, endpoint string) bool {
	mesh := s.meshManager.GetMesh(meshId)

	if mesh == nil {
		return false
	}

	err := mesh.IncrementFailedCount(endpoint)

	if err != nil {
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

func NewSyncErrorHandler(m *mesh.MeshManger) SyncErrorHandler {
	return &SyncErrorHandlerImpl{meshManager: m}
}
