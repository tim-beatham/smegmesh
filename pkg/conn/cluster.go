package conn

import (
	"errors"
	"math"
	"math/rand"
	"slices"
)

// ConnCluster splits nodes into clusters where nodes in a cluster communicate
// frequently and nodes outside of a cluster communicate infrequently
type ConnCluster interface {
	GetNeighbours(global []string, selfId string) []string
	GetInterCluster(global []string, selfId string) string
}

type ConnClusterImpl struct {
	clusterSize int
}

func binarySearch(global []string, selfId string, groupSize int) (int, int) {
	slices.Sort(global)

	lower := 0
	higher := len(global) - 1
	mid := (lower + higher) / 2

	for (higher+1)-lower > groupSize {
		if global[mid] < selfId {
			lower = mid + 1
		} else if global[mid] > selfId {
			higher = mid - 1
		} else {
			break
		}

		mid = (lower + higher) / 2
	}

	return lower, int(math.Min(float64(lower+groupSize), float64(len(global))))
}

// GetNeighbours return the neighbours 'nearest' to you. In this implementation the
// neighbours aren't actually the ones nearest to you but just the ones nearest
// to you alphabetically. Perform binary search to get the total group
func (i *ConnClusterImpl) GetNeighbours(global []string, selfId string) []string {
	slices.Sort(global)

	lower, higher := binarySearch(global, selfId, i.clusterSize)
	// slice the list to get the neighbours
	return global[lower:higher]
}

// GetInterCluster get nodes not in your cluster. Every round there is a given chance
// you will communicate with a random node that is not in your cluster.
func (i *ConnClusterImpl) GetInterCluster(global []string, selfId string) string {
	// Doesn't matter if not in it. Get index of where the node 'should' be
	index, _ := binarySearch(global, selfId, 1)
	numClusters := math.Ceil(float64(len(global)) / float64(i.clusterSize))

	randomCluster := rand.Intn(int(numClusters)-1) + 1

	neighbourIndex := (index + randomCluster) % len(global)
	return global[neighbourIndex]
}

func NewConnCluster(clusterSize int) (ConnCluster, error) {
	log2Cluster := math.Log2(float64(clusterSize))

	if float64((log2Cluster))-log2Cluster != 0 {
		return nil, errors.New("cluster must be a power of 2")
	}

	return &ConnClusterImpl{clusterSize: clusterSize}, nil
}
