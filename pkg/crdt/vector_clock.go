package crdt

import (
	"sync"

	"github.com/tim-beatham/wgmesh/pkg/lib"
)

// Vector clock defines an abstract data type
// for a vector clock implementation
type VectorClock[K comparable] struct {
	vectors   map[K]uint64
	lock      sync.RWMutex
	processID K
}

// IncrementClock: increments the node's value in the vector clock
func (m *VectorClock[K]) IncrementClock() uint64 {
	maxClock := uint64(0)
	m.lock.Lock()

	for _, value := range m.vectors {
		maxClock = max(maxClock, value)
	}

	m.vectors[m.processID] = maxClock + 1
	m.lock.Unlock()
	return maxClock
}

// GetHash: gets the hash of the vector clock used to determine if there
// are any changes
func (m *VectorClock[K]) GetHash() uint64 {
	m.lock.RLock()

	sum := lib.Reduce(uint64(0), lib.MapValues(m.vectors), func(sum uint64, current uint64) uint64 {
		return current + sum
	})

	m.lock.RUnlock()
	return sum
}

func (m *VectorClock[K]) Prune() {
	outliers := lib.GetOutliers(m.vectors, 0.05)

	m.lock.Lock()

	for _, outlier := range outliers {
		delete(m.vectors, outlier)
	}

	m.lock.Unlock()
}

func (m *VectorClock[K]) Put(key K, value uint64) {
	m.lock.Lock()
	m.vectors[key] = max(value, m.vectors[key])
	m.lock.Unlock()
}

func (m *VectorClock[K]) GetClock() map[K]uint64 {
	clock := make(map[K]uint64)

	m.lock.RLock()

	for key, value := range clock {
		clock[key] = value
	}

	m.lock.RUnlock()
	return clock
}

func NewVectorClock[K comparable](processID K) *VectorClock[K] {
	return &VectorClock[K]{
		vectors:   make(map[K]uint64),
		processID: processID,
	}
}
