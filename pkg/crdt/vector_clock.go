package crdt

import (
	"cmp"
	"slices"
	"sync"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/lib"
)

type VectorBucket struct {
	// clock current value of the node's clock
	clock uint64
	// lastUpdate we've seen
	lastUpdate uint64
}

// Vector clock defines an abstract data type
// for a vector clock implementation
type VectorClock[K cmp.Ordered] struct {
	vectors   map[K]*VectorBucket
	lock      sync.RWMutex
	processID K
	staleTime uint64
	hashFunc  func(K) uint64
}

// IncrementClock: increments the node's value in the vector clock
func (m *VectorClock[K]) IncrementClock() uint64 {
	maxClock := uint64(0)
	m.lock.Lock()

	for _, value := range m.vectors {
		maxClock = max(maxClock, value.clock)
	}

	newBucket := VectorBucket{
		clock:      maxClock + 1,
		lastUpdate: uint64(time.Now().Unix()),
	}

	m.vectors[m.processID] = &newBucket

	m.lock.Unlock()
	return maxClock
}

// GetHash: gets the hash of the vector clock used to determine if there
// are any changes
func (m *VectorClock[K]) GetHash() uint64 {
	m.lock.RLock()

	hash := uint64(0)

	sortedKeys := lib.MapKeys(m.vectors)
	slices.Sort(sortedKeys)

	for key, bucket := range m.vectors {
		hash += m.hashFunc(key)
		hash += bucket.clock
	}

	m.lock.RUnlock()
	return hash
}

// getStale: get all entries that are stale within the mesh
func (m *VectorClock[K]) getStale() []K {
	m.lock.RLock()
	maxTimeStamp := lib.Reduce(0, lib.MapValues(m.vectors), func(i uint64, vb *VectorBucket) uint64 {
		return max(i, vb.lastUpdate)
	})

	toRemove := make([]K, 0)

	for key, bucket := range m.vectors {
		if maxTimeStamp-bucket.lastUpdate > m.staleTime {
			toRemove = append(toRemove, key)
		}
	}

	m.lock.RUnlock()
	return toRemove
}

func (m *VectorClock[K]) Prune() {
	stale := m.getStale()

	m.lock.Lock()

	for _, key := range stale {
		delete(m.vectors, key)
	}

	m.lock.Unlock()
}

func (m *VectorClock[K]) GetTimestamp(processId K) uint64 {
	return m.vectors[processId].lastUpdate
}

func (m *VectorClock[K]) Put(key K, value uint64) {
	clockValue := uint64(0)

	m.lock.Lock()
	bucket, ok := m.vectors[key]

	if ok {
		clockValue = bucket.clock
	}

	if value > clockValue {
		newBucket := VectorBucket{
			clock:      value,
			lastUpdate: uint64(time.Now().Unix()),
		}
		m.vectors[key] = &newBucket
	}

	m.lock.Unlock()
}

func (m *VectorClock[K]) GetClock() map[K]uint64 {
	clock := make(map[K]uint64)

	m.lock.RLock()

	keys := lib.MapKeys(m.vectors)
	slices.Sort(keys)

	for key, value := range clock {
		clock[key] = value
	}

	m.lock.RUnlock()
	return clock
}

func NewVectorClock[K cmp.Ordered](processID K, hashFunc func(K) uint64, staleTime uint64) *VectorClock[K] {
	return &VectorClock[K]{
		vectors:   make(map[K]*VectorBucket),
		processID: processID,
		staleTime: staleTime,
		hashFunc:  hashFunc,
	}
}
