package conn

import (
	"math/rand"
	"slices"
	"testing"
)

func TestGetNeighboursClusterSizeTwo(t *testing.T) {
	cluster := &ConnClusterImpl{
		clusterSize: 2,
	}
	neighbours := []string{
		"a",
		"b",
		"c",
		"d",
	}

	result := cluster.GetNeighbours(neighbours, "b")

	if len(result) != 2 {
		t.Fatalf(`neighbour length should be 2`)
	}

	if result[0] != "a" && result[1] != "b" {
		t.Fatalf(`Expected value b`)
	}
}

func TestGetNeighboursGlobalListLessThanClusterSize(t *testing.T) {
	cluster := &ConnClusterImpl{
		clusterSize: 4,
	}

	neighbours := []string{
		"a",
		"b",
		"c",
	}

	result := cluster.GetNeighbours(neighbours, "a")

	if len(result) != 3 {
		t.Fatalf(`neighbour length should be 3`)
	}

	slices.Sort(result)

	if !slices.Equal(result, neighbours) {
		t.Fatalf(`Cluster and neighbours should be equal`)
	}
}

func TestGetNeighboursClusterSize4(t *testing.T) {
	cluster := &ConnClusterImpl{
		clusterSize: 4,
	}

	neighbours := []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o",
	}

	result := cluster.GetNeighbours(neighbours, "k")

	if len(result) != 4 {
		t.Fatalf(`cluster size must be 4`)
	}

	slices.Sort(result)

	if !slices.Equal(neighbours[8:12], result) {
		t.Fatalf(`Cluster should be i, j, k, l`)
	}
}

func TestGetNeighboursClusterSize4OneReturned(t *testing.T) {
	cluster := &ConnClusterImpl{
		clusterSize: 4,
	}

	neighbours := []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o",
	}

	result := cluster.GetNeighbours(neighbours, "o")

	if len(result) != 3 {
		t.Fatalf(`Cluster should be of length 3`)
	}

	if !slices.Equal(neighbours[12:15], result) {
		t.Fatalf(`Cluster should be m, n, o`)
	}
}

func TestInterClusterNotInCluster(t *testing.T) {
	rand.Seed(1)
	cluster := &ConnClusterImpl{
		clusterSize: 4,
	}

	global := []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o",
	}

	neighbours := cluster.GetNeighbours(global, "c")
	interCluster := cluster.GetInterCluster(global, "c")

	if slices.Contains(neighbours, interCluster) {
		t.Fatalf(`intercluster cannot be in your cluster`)
	}
}
