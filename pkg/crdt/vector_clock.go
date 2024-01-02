package crdt

import (
	"cmp"
	"sync"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/lib"
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
	vectors   map[uint64]*VectorBucket
	lock      sync.RWMutex
	processID K
	staleTime uint64
	hashFunc  func(K) uint64
	// highest update that's been garbage collected
	highestStale uint64
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

	m.vectors[m.hashFunc(m.processID)] = &newBucket

	m.lock.Unlock()
	return maxClock
}

// GetHash: gets the hash of the vector clock used to determine if there
// are any changes
func (m *VectorClock[K]) GetHash() uint64 {
	m.lock.RLock()

	hash := uint64(0)

	for key, bucket := range m.vectors {
		hash += key * (bucket.clock + 1)
	}

	m.lock.RUnlock()
	return hash
}

func (m *VectorClock[K]) Merge(vectors map[uint64]uint64) {
	for key, value := range vectors {
		m.put(key, value)
	}
}

// getStale: get all entries that are stale within the mesh
func (m *VectorClock[K]) getStale() []uint64 {
	m.lock.RLock()
	maxTimeStamp := lib.Reduce(0, lib.MapValues(m.vectors), func(i uint64, vb *VectorBucket) uint64 {
		return max(i, vb.lastUpdate)
	})

	toRemove := make([]uint64, 0)

	for key, bucket := range m.vectors {
		if maxTimeStamp-bucket.lastUpdate > m.staleTime {
			toRemove = append(toRemove, key)
			m.highestStale = max(bucket.clock, m.highestStale)
		}
	}

	m.lock.RUnlock()
	return toRemove
}

// GetStaleCount: returns a vector clock which is considered to be stale.
// all updates must be greater than this
func (m *VectorClock[K]) GetStaleCount() uint64 {
	m.lock.RLock()
	staleCount := m.highestStale
	m.lock.RUnlock()
	return staleCount
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
	m.lock.RLock()

	lastUpdate := m.vectors[m.hashFunc(m.processID)].lastUpdate

	m.lock.RUnlock()
	return lastUpdate
}

func (m *VectorClock[K]) Put(key K, value uint64) {
	m.put(m.hashFunc(key), value)
}

func (m *VectorClock[K]) put(key uint64, value uint64) {
	clockValue := uint64(0)

	m.lock.Lock()
	bucket, ok := m.vectors[key]

	if ok {
		clockValue = bucket.clock
	}

	// Make sure that entries that were garbage collected don't get
	// addded back
	if value > clockValue && value > m.highestStale {
		newBucket := VectorBucket{
			clock:      value,
			lastUpdate: uint64(time.Now().Unix()),
		}
		m.vectors[key] = &newBucket
	}

	m.lock.Unlock()
}

func (m *VectorClock[K]) GetClock() map[uint64]uint64 {
	clock := make(map[uint64]uint64)

	m.lock.RLock()

	for key, value := range m.vectors {
		clock[key] = value.clock
	}

	m.lock.RUnlock()
	return clock
}

func NewVectorClock[K cmp.Ordered](processID K, hashFunc func(K) uint64, staleTime uint64) *VectorClock[K] {
	return &VectorClock[K]{
		vectors:   make(map[uint64]*VectorBucket),
		processID: processID,
		staleTime: staleTime,
		hashFunc:  hashFunc,
	}
}
