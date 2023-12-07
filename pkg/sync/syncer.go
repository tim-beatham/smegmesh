package sync

import (
	"io"
	"math/rand"
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
		randomSubset[len(randomSubset)-1] = s.cluster.GetInterCluster(nodeNames, publicKey.String())
	}

	var succeeded bool = false

	// Do this synchronously to conserve bandwidth
	for _, node := range randomSubset {
		correspondingPeer := s.manager.GetNode(meshId, node)

		if correspondingPeer == nil {
			logging.Log.WriteErrorf("node %s does not exist", node)
		}

		err = s.requester.SyncMesh(meshId, correspondingPeer)

		if err == nil || err == io.EOF {
			succeeded = true
		} else {
			// If the synchronisation operation has failed them mark a gravestone
			// preventing the peer from being re-contacted until it has updated
			// itself
			s.manager.GetMesh(meshId).Mark(node)
		}
	}

	s.syncCount++
	logging.Log.WriteInfof("SYNC TIME: %v", time.Since(before))
	logging.Log.WriteInfof("SYNC COUNT: %d", s.syncCount)

	s.infectionCount = ((s.conf.InfectionCount + s.infectionCount - 1) % s.conf.InfectionCount)

	if !succeeded {
		// If could not gossip with anyone then repeat.
		s.infectionCount++
	}

	s.manager.GetMesh(meshId).SaveChanges()
	s.manager.Prune()
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
