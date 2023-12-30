package sync

import (
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

// Syncer: picks random nodes from the meshs
type Syncer interface {
	Sync(theMesh mesh.MeshProvider) error
	SyncMeshes() error
}

type SyncerImpl struct {
	manager        mesh.MeshManager
	requester      SyncRequester
	infectionCount int
	syncCount      int
	cluster        conn.ConnCluster
	conf           *conf.DaemonConfiguration
	lastSync       map[string]uint64
}

// Sync: Sync with random nodes
func (s *SyncerImpl) Sync(correspondingMesh mesh.MeshProvider) error {
	if correspondingMesh == nil {
		return fmt.Errorf("mesh provided was nil cannot sync nil mesh")
	}

	// Self can be nil if the node is removed
	selfID := s.manager.GetPublicKey()
	self, _ := correspondingMesh.GetNode(selfID.String())

	// Mesh has been removed
	if self == nil {
		return fmt.Errorf("mesh %s does not exist", correspondingMesh.GetMeshId())
	}

	correspondingMesh.Prune()

	if correspondingMesh.HasChanges() {
		logging.Log.WriteInfof("meshes %s has changes", correspondingMesh.GetMeshId())
	}

	if self.GetType() == conf.PEER_ROLE && !correspondingMesh.HasChanges() && s.infectionCount == 0 {
		logging.Log.WriteInfof("no changes for %s", correspondingMesh.GetMeshId())

		// If not synchronised in certain pull from random neighbour
		if uint64(time.Now().Unix())-s.lastSync[correspondingMesh.GetMeshId()] > 20 {
			return s.Pull(self, correspondingMesh)
		}

		return nil
	}

	before := time.Now()
	s.manager.GetRouteManager().UpdateRoutes()

	publicKey := s.manager.GetPublicKey()
	nodeNames := correspondingMesh.GetPeers()

	if self != nil {
		nodeNames = lib.Filter(nodeNames, func(s string) bool {
			return s != mesh.NodeID(self)
		})
	}

	var gossipNodes []string

	// Clients always pings its peer for configuration
	if self != nil && self.GetType() == conf.CLIENT_ROLE && len(nodeNames) > 1 {
		neighbours := s.cluster.GetNeighbours(nodeNames, publicKey.String())

		if len(neighbours) == 0 {
			return nil
		}

		redundancyLength := min(len(neighbours), 3)
		gossipNodes = neighbours[:redundancyLength]
	} else {
		neighbours := s.cluster.GetNeighbours(nodeNames, publicKey.String())
		gossipNodes = lib.RandomSubsetOfLength(neighbours, s.conf.BranchRate)

		if len(nodeNames) > s.conf.ClusterSize && rand.Float64() < s.conf.InterClusterChance {
			gossipNodes[len(gossipNodes)-1] = s.cluster.GetInterCluster(nodeNames, publicKey.String())
		}
	}

	var succeeded bool = false

	// Do this synchronously to conserve bandwidth
	for _, node := range gossipNodes {
		correspondingPeer := s.manager.GetNode(correspondingMesh.GetMeshId(), node)

		if correspondingPeer == nil {
			logging.Log.WriteErrorf("node %s does not exist", node)
			continue
		}

		err := s.requester.SyncMesh(correspondingMesh.GetMeshId(), correspondingPeer)

		if err == nil || err == io.EOF {
			succeeded = true
		}

		if err != nil {
			logging.Log.WriteInfof(err.Error())
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

	correspondingMesh.SaveChanges()

	s.lastSync[correspondingMesh.GetMeshId()] = uint64(time.Now().Unix())
	return nil
}

// Pull one node in the cluster, if there has not been message dissemination
// in a certain period of time pull a random node within the cluster
func (s *SyncerImpl) Pull(self mesh.MeshNode, mesh mesh.MeshProvider) error {
	peers := mesh.GetPeers()
	pubKey, _ := self.GetPublicKey()

	neighbours := s.cluster.GetNeighbours(peers, pubKey.String())
	neighbour := lib.RandomSubsetOfLength(neighbours, 1)

	if len(neighbour) == 0 {
		logging.Log.WriteInfof("no neighbours")
		return nil
	}

	logging.Log.WriteInfof("PULLING from node %s", neighbour[0])

	pullNode, err := mesh.GetNode(neighbour[0])

	if err != nil || pullNode == nil {
		return fmt.Errorf("node %s does not exist in the mesh", neighbour[0])
	}

	err = s.requester.SyncMesh(mesh.GetMeshId(), pullNode)

	if err == nil || err == io.EOF {
		s.lastSync[mesh.GetMeshId()] = uint64(time.Now().Unix())
	} else {
		return err
	}

	s.syncCount++
	return nil
}

// SyncMeshes: Sync all meshes
func (s *SyncerImpl) SyncMeshes() error {
	var wg sync.WaitGroup

	for _, mesh := range s.manager.GetMeshes() {
		wg.Add(1)

		sync := func() {
			defer wg.Done()

			err := s.Sync(mesh)

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
			}
		}

		go sync()
	}

	logging.Log.WriteInfof("updating the WireGuard configuration")
	err := s.manager.ApplyConfig()

	if err != nil {
		logging.Log.WriteInfof("failed to update config %w", err)
	}
	return nil
}

func NewSyncer(m mesh.MeshManager, conf *conf.DaemonConfiguration, r SyncRequester) Syncer {
	cluster, _ := conn.NewConnCluster(conf.ClusterSize)
	return &SyncerImpl{
		manager:        m,
		conf:           conf,
		requester:      r,
		infectionCount: 0,
		syncCount:      0,
		cluster:        cluster,
		lastSync:       make(map[string]uint64)}
}
