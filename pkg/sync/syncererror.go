package sync

import (
	"github.com/tim-beatham/smegmesh/pkg/conn"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SyncErrorHandler: Handles errors when attempting to sync
type SyncErrorHandler interface {
	Handle(mesh mesh.MeshProvider, endpoint string, err error) bool
}

// SyncErrorHandlerImpl Is an implementation of the SyncErrorHandler
type SyncErrorHandlerImpl struct {
	meshManager mesh.MeshManager
	connManager conn.ConnectionManager
}

func (s *SyncErrorHandlerImpl) handleFailed(mesh mesh.MeshProvider, nodeId string) bool {
	mesh.Mark(nodeId)
	node, err := mesh.GetNode(nodeId)

	if err != nil {
		s.connManager.RemoveConnection(node.GetHostEndpoint())
	}
	return true
}

func (s *SyncErrorHandlerImpl) handleDeadlineExceeded(mesh mesh.MeshProvider, nodeId string) bool {
	node, err := mesh.GetNode(nodeId)

	if err != nil {
		return false
	}

	s.connManager.RemoveConnection(node.GetHostEndpoint())
	return true
}

func (s *SyncErrorHandlerImpl) Handle(mesh mesh.MeshProvider, nodeId string, err error) bool {
	errStatus, _ := status.FromError(err)

	logging.Log.WriteInfof("Handled gRPC error: %s", errStatus.Message())

	switch errStatus.Code() {
	case codes.Unavailable, codes.Unknown, codes.Internal, codes.NotFound:
		return s.handleFailed(mesh, nodeId)
	case codes.DeadlineExceeded:
		return s.handleDeadlineExceeded(mesh, nodeId)
	}

	return false
}

func NewSyncErrorHandler(m mesh.MeshManager, conn conn.ConnectionManager) SyncErrorHandler {
	return &SyncErrorHandlerImpl{meshManager: m, connManager: conn}
}
