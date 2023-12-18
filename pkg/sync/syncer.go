package sync

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

// Syncer: picks random nodes from the meshs
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
	conf           *conf.DaemonConfiguration
	lastSync       map[string]uint64
}

// Sync: Sync random nodes
func (s *SyncerImpl) Sync(meshId string) error {
	// Self can be nil if the node is removed
	self, _ := s.manager.GetSelf(meshId)

	correspondingMesh := s.manager.GetMesh(meshId)

	correspondingMesh.Prune()

	if self != nil && self.GetType() == conf.PEER_ROLE && !s.manager.HasChanges(meshId) && s.infectionCount == 0 {
		logging.Log.WriteInfof("No changes for %s", meshId)

		// If not synchronised in certain pull from random neighbour
		if uint64(time.Now().Unix())-s.lastSync[meshId] > 20 {
			return s.Pull(meshId)
		}

		return nil
	}

	before := time.Now()
	s.manager.GetRouteManager().UpdateRoutes()

	publicKey := s.manager.GetPublicKey()

	logging.Log.WriteInfof(publicKey.String())

	nodeNames := correspondingMesh.GetPeers()

	if self != nil {
		nodeNames = lib.Filter(nodeNames, func(s string) bool {
			return s != mesh.NodeID(self)
		})
	}

	var gossipNodes []string

	// Clients always pings its peer for configuration
	if self != nil && self.GetType() == conf.CLIENT_ROLE {
		keyFunc := lib.HashString
		bucketFunc := lib.HashString

		neighbour := lib.ConsistentHash(nodeNames, publicKey.String(), keyFunc, bucketFunc)
		gossipNodes = make([]string, 1)
		gossipNodes[0] = neighbour
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
		correspondingPeer := s.manager.GetNode(meshId, node)

		if correspondingPeer == nil {
			logging.Log.WriteErrorf("node %s does not exist", node)
		}

		err := s.requester.SyncMesh(meshId, correspondingPeer)

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
	s.lastSync[meshId] = uint64(time.Now().Unix())

	logging.Log.WriteInfof("UPDATING WG CONF")
	err := s.manager.ApplyConfig()

	if err != nil {
		logging.Log.WriteInfof("Failed to update config %w", err)
	}

	return nil
}

// Pull one node in the cluster, if there has not been message dissemination
// in a certain period of time pull a random node within the cluster
func (s *SyncerImpl) Pull(meshId string) error {
	mesh := s.manager.GetMesh(meshId)
	self, err := s.manager.GetSelf(meshId)

	if err != nil {
		return err
	}

	pubKey, _ := self.GetPublicKey()

	if mesh == nil {
		return errors.New("mesh is nil, invalid operation")
	}

	peers := mesh.GetPeers()
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

	err = s.requester.SyncMesh(meshId, pullNode)

	if err == nil || err == io.EOF {
		s.lastSync[meshId] = uint64(time.Now().Unix())
	} else {
		return err
	}

	s.syncCount++
	return nil
}

// SyncMeshes: Sync all meshes
func (s *SyncerImpl) SyncMeshes() error {
	for meshId := range s.manager.GetMeshes() {
		err := s.Sync(meshId)

		if err != nil {
			logging.Log.WriteErrorf(err.Error())
		}
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
