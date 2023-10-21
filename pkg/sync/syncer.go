package sync

import (
	"errors"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/manager"
)

// Syncer: picks random nodes from the mesh
type Syncer interface {
	Sync(meshId string) error
	SyncMeshes() error
}

type SyncerImpl struct {
	manager            *manager.MeshManger
	requester          SyncRequester
	authenticatedNodes []crdt.MeshNodeCrdt
}

const subSetLength = 5
const maxAuthentications = 30

// Sync: Sync random nodes
func (s *SyncerImpl) Sync(meshId string) error {
	mesh := s.manager.GetMesh(meshId)

	if mesh == nil {
		return errors.New("the provided mesh does not exist")
	}

	snapshot, err := mesh.GetCrdt()

	if err != nil {
		return err
	}

	if len(snapshot.Nodes) <= 1 {
		return nil
	}

	excludedNodes := map[string]struct{}{
		s.manager.HostEndpoint: {},
	}

	meshNodes := lib.MapValuesWithExclude(snapshot.Nodes, excludedNodes)
	randomSubset := lib.RandomSubsetOfLength(meshNodes, subSetLength)

	for _, n := range randomSubset {
		err := s.requester.SyncMesh(meshId, n.HostEndpoint)

		if err != nil {
			return err
		}
	}

	return nil
}

// SyncMeshes: Sync all meshes
func (s *SyncerImpl) SyncMeshes() error {
	for _, m := range s.manager.Meshes {
		err := s.Sync(m.MeshId)

		if err != nil {
			return err
		}
	}

	return s.manager.ApplyWg()
}

func NewSyncer(m *manager.MeshManger, r SyncRequester) Syncer {
	return &SyncerImpl{manager: m, requester: r}
}
