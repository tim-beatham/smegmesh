package sync

import (
	"errors"
	"sync"
	"time"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

// Syncer: picks random nodes from the mesh
type Syncer interface {
	Sync(meshId string) error
	SyncMeshes() error
}

type SyncerImpl struct {
	manager            *mesh.MeshManger
	requester          SyncRequester
	authenticatedNodes []crdt.MeshNodeCrdt
}

const subSetLength = 5
const maxAuthentications = 30

// Sync: Sync random nodes
func (s *SyncerImpl) Sync(meshId string) error {
	if !s.manager.HasChanges(meshId) {
		logging.Log.WriteInfof("No changes for %s", meshId)
		return nil
	}

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

	for _, node := range snapshot.Nodes {
		if mesh.HasFailed(node.HostEndpoint) {
			excludedNodes[node.HostEndpoint] = struct{}{}
		}
	}

	meshNodes := lib.MapValuesWithExclude(snapshot.Nodes, excludedNodes)
	randomSubset := lib.RandomSubsetOfLength(meshNodes, subSetLength)

	before := time.Now()

	var waitGroup sync.WaitGroup

	for _, n := range randomSubset {
		waitGroup.Add(1)

		syncMeshFunc := func() error {
			defer waitGroup.Done()
			err := s.requester.SyncMesh(meshId, n.HostEndpoint)
			return err
		}

		go syncMeshFunc()
	}

	waitGroup.Wait()

	logging.Log.WriteInfof("SYNC TIME: %v", time.Now().Sub(before))
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

	return nil
}

func NewSyncer(m *mesh.MeshManger, r SyncRequester) Syncer {
	return &SyncerImpl{manager: m, requester: r}
}
