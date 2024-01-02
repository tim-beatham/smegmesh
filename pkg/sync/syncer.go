package sync

import (
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/conn"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
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
	lastSync       map[string]int64
}

// Sync: Sync with random nodes
func (s *SyncerImpl) Sync(correspondingMesh mesh.MeshProvider) error {
	if correspondingMesh == nil {
		return fmt.Errorf("mesh provided was nil cannot sync nil mesh")
	}

	// Self can be nil if the node is removed
	selfID := s.manager.GetPublicKey()
	self, err := correspondingMesh.GetNode(selfID.String())

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
	}

	correspondingMesh.Prune()

	if correspondingMesh.HasChanges() {
		logging.Log.WriteInfof("meshes %s has changes", correspondingMesh.GetMeshId())
	}

	// If removed sync with other nodes to gossip the node is removed
	if self != nil && self.GetType() == conf.PEER_ROLE && !correspondingMesh.HasChanges() && s.infectionCount == 0 {
		logging.Log.WriteInfof("no changes for %s", correspondingMesh.GetMeshId())

		// If not synchronised in certain time pull from random neighbour
		if s.conf.PullTime != 0 && time.Now().Unix()-s.lastSync[correspondingMesh.GetMeshId()] > int64(s.conf.PullTime) {
			return s.Pull(self, correspondingMesh)
		}

		return nil
	}

	before := time.Now()
	err = s.manager.GetRouteManager().UpdateRoutes()

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
	}

	publicKey := s.manager.GetPublicKey()
	nodeNames := correspondingMesh.GetPeers()

	nodeNames = lib.Filter(nodeNames, func(s string) bool {
		// Filter our only public key out so we dont sync with ourself
		return s != publicKey.String()
	})

	var gossipNodes []string

	// Clients always pings its peer for configuration
	if self != nil && self.GetType() == conf.CLIENT_ROLE && len(nodeNames) > 1 {
		neighbours := s.cluster.GetNeighbours(nodeNames, publicKey.String())

		if len(neighbours) == 0 {
			return nil
		}

		// Peer with 2 nodes so that there is redundnacy in
		// the situation the node leaves pre-emptively
		redundancyLength := min(len(neighbours), 2)
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
		correspondingPeer, err := correspondingMesh.GetNode(node)

		if correspondingPeer == nil || err != nil {
			logging.Log.WriteErrorf("node %s does not exist", node)
			continue
		}

		err = s.requester.SyncMesh(correspondingMesh.GetMeshId(), correspondingPeer)

		if err == nil || err == io.EOF {
			succeeded = true
		}

		if err != nil {
			logging.Log.WriteErrorf(err.Error())
		}
	}

	s.syncCount++
	logging.Log.WriteInfof("sync time: %v", time.Since(before))
	logging.Log.WriteInfof("number of syncs: %d", s.syncCount)

	s.infectionCount = ((s.conf.InfectionCount + s.infectionCount - 1) % s.conf.InfectionCount)

	if !succeeded {
		s.infectionCount++
	}

	correspondingMesh.SaveChanges()

	s.lastSync[correspondingMesh.GetMeshId()] = time.Now().Unix()
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

	logging.Log.WriteInfof("pulling from node %s", neighbour[0])

	pullNode, err := mesh.GetNode(neighbour[0])

	if err != nil || pullNode == nil {
		return fmt.Errorf("node %s does not exist in the mesh", neighbour[0])
	}

	err = s.requester.SyncMesh(mesh.GetMeshId(), pullNode)

	if err == nil || err == io.EOF {
		s.lastSync[mesh.GetMeshId()] = time.Now().Unix()
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
		lastSync:       make(map[string]int64)}
}
