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
	manager            *mesh.MeshManager
	requester          SyncRequester
	authenticatedNodes []crdt.MeshNodeCrdt
	infectionCount     int
}

const subSetLength = 3
const infectionCount = 3

// Sync: Sync random nodes
func (s *SyncerImpl) Sync(meshId string) error {
	logging.Log.WriteInfof("UPDATING WG CONF")
	s.manager.ApplyConfig()

	if !s.manager.HasChanges(meshId) && s.infectionCount == 0 {
		logging.Log.WriteInfof("No changes for %s", meshId)
		return nil
	}

	mesh := s.manager.GetMesh(meshId)

	if mesh == nil {
		return errors.New("the provided mesh does not exist")
	}

	snapshot, err := mesh.GetMesh()

	if err != nil {
		return err
	}

	nodes := snapshot.GetNodes()

	if len(nodes) <= 1 {
		return nil
	}

	excludedNodes := map[string]struct{}{
		s.manager.HostParameters.HostEndpoint: {},
	}

	meshNodes := lib.MapValuesWithExclude(nodes, excludedNodes)
	randomSubset := lib.RandomSubsetOfLength(meshNodes, subSetLength)

	before := time.Now()

	var waitGroup sync.WaitGroup

	for _, n := range randomSubset {
		waitGroup.Add(1)

		syncMeshFunc := func() error {
			defer waitGroup.Done()
			err := s.requester.SyncMesh(meshId, n.GetHostEndpoint())
			return err
		}

		go syncMeshFunc()
	}

	waitGroup.Wait()

	logging.Log.WriteInfof("SYNC TIME: %v", time.Now().Sub(before))

	s.infectionCount = ((infectionCount + s.infectionCount - 1) % infectionCount)

	return nil
}

// SyncMeshes: Sync all meshes
func (s *SyncerImpl) SyncMeshes() error {
	for meshId, _ := range s.manager.Meshes {
		err := s.Sync(meshId)

		if err != nil {
			return err
		}
	}

	return nil
}

func NewSyncer(m *mesh.MeshManager, r SyncRequester) Syncer {
	return &SyncerImpl{manager: m, requester: r, infectionCount: 0}
}
