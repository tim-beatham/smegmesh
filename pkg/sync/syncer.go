package sync

import (
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
	manager        mesh.MeshManager
	requester      SyncRequester
	infectionCount int
	syncCount      int
	cluster        conn.ConnCluster
	conf           *conf.WgMeshConfiguration
}

// Sync: Sync random nodes
func (s *SyncerImpl) Sync(meshId string) error {
	if !s.manager.HasChanges(meshId) && s.infectionCount == 0 {
		logging.Log.WriteInfof("No changes for %s", meshId)
		return nil
	}

	logging.Log.WriteInfof("UPDATING WG CONF")

	s.manager.GetRouteManager().UpdateRoutes()
	err := s.manager.ApplyConfig()

	if err != nil {
		logging.Log.WriteInfof("Failed to update config %w", err)
	}

	publicKey := s.manager.GetPublicKey()

	logging.Log.WriteInfof(publicKey.String())

	nodeNames := s.manager.GetMesh(meshId).GetPeers()
	neighbours := s.cluster.GetNeighbours(nodeNames, publicKey.String())
	randomSubset := lib.RandomSubsetOfLength(neighbours, s.conf.BranchRate)

	for _, node := range randomSubset {
		logging.Log.WriteInfof("Random node: %s", node)
	}

	before := time.Now()

	if len(nodeNames) > s.conf.ClusterSize && rand.Float64() < s.conf.InterClusterChance {
		logging.Log.WriteInfof("Sending to random cluster")
		interCluster := s.cluster.GetInterCluster(nodeNames, publicKey.String())
		randomSubset = append(randomSubset, interCluster)
	}

	var waitGroup sync.WaitGroup

	for index := range randomSubset {
		waitGroup.Add(1)

		go func(i int) error {
			defer waitGroup.Done()

			correspondingPeer := s.manager.GetNode(meshId, randomSubset[i])

			if correspondingPeer == nil {
				logging.Log.WriteErrorf("node %s does not exist", randomSubset[i])
			}

			err := s.requester.SyncMesh(meshId, correspondingPeer.GetHostEndpoint())
			return err
		}(index)
	}

	waitGroup.Wait()

	s.syncCount++
	logging.Log.WriteInfof("SYNC TIME: %v", time.Since(before))
	logging.Log.WriteInfof("SYNC COUNT: %d", s.syncCount)

	s.infectionCount = ((s.conf.InfectionCount + s.infectionCount - 1) % s.conf.InfectionCount)
	s.manager.GetMesh(meshId).SaveChanges()
	return nil
}

// SyncMeshes: Sync all meshes
func (s *SyncerImpl) SyncMeshes() error {
	for meshId := range s.manager.GetMeshes() {
		err := s.Sync(meshId)

		if err != nil {
			return err
		}
	}

	return nil
}

func NewSyncer(m mesh.MeshManager, conf *conf.WgMeshConfiguration, r SyncRequester) Syncer {
	cluster, _ := conn.NewConnCluster(conf.ClusterSize)
	return &SyncerImpl{
		manager:        m,
		conf:           conf,
		requester:      r,
		infectionCount: 0,
		syncCount:      0,
		cluster:        cluster}
}
