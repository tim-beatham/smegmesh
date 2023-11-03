package sync

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
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
	manager        *mesh.MeshManager
	requester      SyncRequester
	infectionCount int
	syncCount      int
	cluster        conn.ConnCluster
	conf           *conf.WgMeshConfiguration
}

// Sync: Sync random nodes
func (s *SyncerImpl) Sync(meshId string) error {
	logging.Log.WriteInfof("UPDATING WG CONF")
	s.manager.ApplyConfig()

	if !s.manager.HasChanges(meshId) && s.infectionCount == 0 {
		logging.Log.WriteInfof("No changes for %s", meshId)
		return nil
	}

	theMesh := s.manager.GetMesh(meshId)

	if theMesh == nil {
		return errors.New("the provided mesh does not exist")
	}

	snapshot, err := theMesh.GetMesh()

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

	getNames := func(node mesh.MeshNode) string {
		return node.GetHostEndpoint()
	}

	nodeNames := lib.Map(meshNodes, getNames)

	neighbours := s.cluster.GetNeighbours(nodeNames, s.manager.HostParameters.HostEndpoint)
	randomSubset := lib.RandomSubsetOfLength(neighbours, s.conf.BranchRate)

	for _, node := range randomSubset {
		logging.Log.WriteInfof("Random node: %s", node)
	}

	before := time.Now()

	if len(meshNodes) > s.conf.ClusterSize && rand.Float64() < s.conf.InterClusterChance {
		logging.Log.WriteInfof("Sending to random cluster")
		interCluster := s.cluster.GetInterCluster(nodeNames, s.manager.HostParameters.HostEndpoint)
		randomSubset = append(randomSubset, interCluster)
	}

	var waitGroup sync.WaitGroup

	for index := range randomSubset {
		waitGroup.Add(1)

		go func(i int) error {
			defer waitGroup.Done()
			err := s.requester.SyncMesh(meshId, randomSubset[i])
			return err
		}(index)
	}

	waitGroup.Wait()

	s.syncCount++
	logging.Log.WriteInfof("SYNC TIME: %v", time.Now().Sub(before))
	logging.Log.WriteInfof("SYNC COUNT: %d", s.syncCount)

	s.infectionCount = ((s.conf.InfectionCount + s.infectionCount - 1) % s.conf.InfectionCount)

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

func NewSyncer(m *mesh.MeshManager, conf *conf.WgMeshConfiguration, r SyncRequester) Syncer {
	cluster, _ := conn.NewConnCluster(conf.ClusterSize)
	return &SyncerImpl{
		manager:        m,
		conf:           conf,
		requester:      r,
		infectionCount: 0,
		syncCount:      0,
		cluster:        cluster}
}
