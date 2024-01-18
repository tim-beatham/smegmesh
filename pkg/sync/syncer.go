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
	Sync(theMesh mesh.MeshProvider) (bool, error)
	SyncMeshes() error
}

// SyncerImpl: implementation of a syncer to sync meshes
type SyncerImpl struct {
	meshManager    mesh.MeshManager
	requester      SyncRequester
	infectionCount int
	syncCount      int
	cluster        conn.ConnCluster
	configuration  *conf.DaemonConfiguration
	lastSync       map[string]int64
	lastPoll       map[string]int64
	lastSyncLock   sync.RWMutex
	lastPollLock   sync.RWMutex
}

// Sync: Sync with random nodes. Returns true if there was changes false otherwise
func (s *SyncerImpl) Sync(correspondingMesh mesh.MeshProvider) (bool, error) {
	if correspondingMesh == nil {
		return false, fmt.Errorf("mesh provided was nil cannot sync nil mesh")
	}

	// Self can be nil if the node is removed
	selfID := s.meshManager.GetPublicKey()
	self, _ := correspondingMesh.GetNode(selfID.String())

	correspondingMesh.Prune()

	if correspondingMesh.HasChanges() {
		logging.Log.WriteInfof("meshes %s has changes", correspondingMesh.GetMeshId())
	}

	// If removed sync with other nodes to gossip the node is removed
	if self != nil && self.GetType() == conf.PEER_ROLE && !correspondingMesh.HasChanges() && s.infectionCount == 0 {
		logging.Log.WriteInfof("no changes for %s", correspondingMesh.GetMeshId())

		// If not synchronised in certain time pull from random neighbour
		if s.configuration.PullInterval != 0 && time.Now().Unix()-s.lastSync[correspondingMesh.GetMeshId()] > int64(s.configuration.PullInterval) {
			return s.Pull(self, correspondingMesh)
		}

		return false, nil
	}

	before := time.Now()

	publicKey := s.meshManager.GetPublicKey()
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
			return false, nil
		}

		// Peer with 2 nodes so that there is redundnacy in
		// the situation the node leaves pre-emptively
		redundancyLength := min(len(neighbours), 2)
		gossipNodes = neighbours[:redundancyLength]
	} else {
		neighbours := s.cluster.GetNeighbours(nodeNames, publicKey.String())
		gossipNodes = lib.RandomSubsetOfLength(neighbours, s.configuration.Branch)

		if len(nodeNames) > s.configuration.ClusterSize && rand.Float64() < s.configuration.InterClusterChance {
			gossipNodes[len(gossipNodes)-1] = s.cluster.GetInterCluster(nodeNames, publicKey.String())
		}
	}

	var succeeded bool = false

	for _, node := range gossipNodes {
		correspondingPeer, err := correspondingMesh.GetNode(node)

		if correspondingPeer == nil || err != nil {
			logging.Log.WriteErrorf("node %s does not exist", node)
			continue
		}

		err = s.requester.SyncMesh(correspondingMesh, correspondingPeer)

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

	s.infectionCount = ((s.configuration.InfectionCount + s.infectionCount - 1) % s.configuration.InfectionCount)

	if !succeeded {
		s.infectionCount++
	}

	changes := correspondingMesh.HasChanges()
	correspondingMesh.SaveChanges()

	s.lastSyncLock.Lock()
	s.lastSync[correspondingMesh.GetMeshId()] = time.Now().Unix()
	s.lastSyncLock.Unlock()
	return changes, nil
}

// Pull one node in the cluster, if there has not been message dissemination
// in a certain period of time pull a random node within the cluster
func (s *SyncerImpl) Pull(self mesh.MeshNode, mesh mesh.MeshProvider) (bool, error) {
	peers := mesh.GetPeers()
	pubKey, _ := self.GetPublicKey()

	neighbours := s.cluster.GetNeighbours(peers, pubKey.String())
	neighbour := lib.RandomSubsetOfLength(neighbours, 1)

	if len(neighbour) == 0 {
		logging.Log.WriteInfof("no neighbours")
		return false, nil
	}

	logging.Log.WriteInfof("pulling from node %s", neighbour[0])

	pullNode, err := mesh.GetNode(neighbour[0])

	if err != nil || pullNode == nil {
		return false, fmt.Errorf("node %s does not exist in the mesh", neighbour[0])
	}

	err = s.requester.SyncMesh(mesh, pullNode)

	if err == nil || err == io.EOF {
		s.lastSync[mesh.GetMeshId()] = time.Now().Unix()
	} else {
		return false, err
	}

	s.syncCount++

	changes := mesh.HasChanges()
	return changes, nil
}

// SyncMeshes: Sync all meshes
func (s *SyncerImpl) SyncMeshes() error {
	var wg sync.WaitGroup

	meshes := s.meshManager.GetMeshes()

	s.lastPollLock.Lock()
	meshesToSync := lib.Filter(lib.MapValues(meshes), func(mesh mesh.MeshProvider) bool {
		return time.Now().Unix()-s.lastPoll[mesh.GetMeshId()] >= int64(s.configuration.SyncInterval)
	})
	s.lastPollLock.Unlock()

	changes := make(chan bool, len(meshesToSync))

	for i := 0; i < len(meshesToSync); {
		wg.Add(1)

		sync := func(index int) {
			defer wg.Done()

			var hasChanges bool = false

			mesh := meshesToSync[index]

			hasChanges, err := s.Sync(mesh)
			changes <- hasChanges

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
			}

			s.lastPollLock.Lock()
			s.lastPoll[mesh.GetMeshId()] = time.Now().Unix()
			s.lastPollLock.Unlock()
		}

		go sync(i)
		i++
	}
	wg.Wait()

	hasChanges := false

	for i := 0; i < len(changes); i++ {
		if <-changes {
			hasChanges = true
		}
	}

	var err error

	err = s.meshManager.GetRouteManager().UpdateRoutes()
	if err != nil {
		logging.Log.WriteErrorf("update routes failed %s", err.Error())
	}

	if hasChanges {
		logging.Log.WriteInfof("updating the WireGuard configuration")
		err = s.meshManager.ApplyConfig()

		if err != nil {
			logging.Log.WriteErrorf("failed to update config %s", err.Error())
		}
	}

	return nil
}

type NewSyncerParams struct {
	MeshManager       mesh.MeshManager
	ConnectionManager conn.ConnectionManager
	Configuration     *conf.DaemonConfiguration
	Requester         SyncRequester
}

func NewSyncer(params *NewSyncerParams) Syncer {
	cluster, _ := conn.NewConnCluster(params.Configuration.ClusterSize)
	syncRequester := NewSyncRequester(NewSyncRequesterParams{
		MeshManager:       params.MeshManager,
		ConnectionManager: params.ConnectionManager,
		Configuration:     params.Configuration,
	})

	return &SyncerImpl{
		meshManager:    params.MeshManager,
		configuration:  params.Configuration,
		requester:      syncRequester,
		infectionCount: 0,
		syncCount:      0,
		cluster:        cluster,
		lastSync:       make(map[string]int64),
		lastPoll:       make(map[string]int64)}
}
