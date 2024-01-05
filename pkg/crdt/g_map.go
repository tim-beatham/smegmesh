// crdt provides go implementations for crdts
package crdt

import (
	"cmp"
	"sync"
)

// Bucket: bucket represents a value in the grow only map
type Bucket[D any] struct {
	Vector     uint64
	Contents   D
	Gravestone bool
}

// GMap is a set that can only grow in size
type GMap[K cmp.Ordered, D any] struct {
	lock     sync.RWMutex
	contents map[uint64]Bucket[D]
	clock    *VectorClock[K]
}

// Put: put a new entry in the grow-only-map
func (g *GMap[K, D]) Put(key K, value D) {
	g.lock.Lock()

	clock := g.clock.IncrementClock()

	g.contents[g.clock.hashFunc(key)] = Bucket[D]{
		Vector:   clock,
		Contents: value,
	}

	g.lock.Unlock()
}

// Contains: returns whether or not the key is contained
// in the g-map
func (g *GMap[K, D]) Contains(key K) bool {
	return g.contains(g.clock.hashFunc(key))
}

func (g *GMap[K, D]) contains(key uint64) bool {
	g.lock.RLock()

	_, ok := g.contents[key]

	g.lock.RUnlock()

	return ok
}

func (g *GMap[K, D]) put(key uint64, b Bucket[D]) {
	g.lock.Lock()

	if g.contents[key].Vector < b.Vector {
		g.contents[key] = b
	}

	g.lock.Unlock()
}

func (g *GMap[K, D]) get(key uint64) Bucket[D] {
	g.lock.RLock()
	bucket := g.contents[key]
	g.lock.RUnlock()

	return bucket
}

// Get: get the value associated with the given key
func (g *GMap[K, D]) Get(key K) D {
	if !g.Contains(key) {
		var def D
		return def
	}

	return g.get(g.clock.hashFunc(key)).Contents
}

// Mark: marks the node, this means the status of the node
// is an undefined state
func (g *GMap[K, D]) Mark(key K) {
	if !g.Contains(key) {
		return
	}

	g.lock.Lock()
	bucket := g.contents[g.clock.hashFunc(key)]
	bucket.Gravestone = true
	g.contents[g.clock.hashFunc(key)] = bucket
	g.lock.Unlock()
}

// IsMarked: returns true if the node is marked (in an undefined state)
func (g *GMap[K, D]) IsMarked(key K) bool {
	marked := false

	g.lock.RLock()

	bucket, ok := g.contents[g.clock.hashFunc(key)]

	if ok {
		marked = bucket.Gravestone
	}

	g.lock.RUnlock()
	return marked
}

// Keys: return all the keys in the grow-only map
func (g *GMap[K, D]) Keys() []uint64 {
	g.lock.RLock()

	contents := make([]uint64, len(g.contents))
	index := 0

	for key := range g.contents {
		contents[index] = key
		index++
	}

	g.lock.RUnlock()
	return contents
}

// Save: saves the grow only map
func (g *GMap[K, D]) Save() map[uint64]Bucket[D] {
	buckets := make(map[uint64]Bucket[D])
	g.lock.RLock()

	for key, value := range g.contents {
		buckets[key] = value
	}

	g.lock.RUnlock()
	return buckets
}

// SaveWithKeys: get all the values corresponding with the provided keys
func (g *GMap[K, D]) SaveWithKeys(keys []uint64) map[uint64]Bucket[D] {
	buckets := make(map[uint64]Bucket[D])
	g.lock.RLock()

	for _, key := range keys {
		buckets[key] = g.contents[key]
	}

	g.lock.RUnlock()
	return buckets
}

// GetClock: get all the vector clocks in the g_map
func (g *GMap[K, D]) GetClock() map[uint64]uint64 {
	clock := make(map[uint64]uint64)
	g.lock.RLock()

	for key, bucket := range g.contents {
		clock[key] = bucket.Vector
	}

	g.lock.RUnlock()
	return clock
}

// GetHash: get the hash of the g_map representing its state
func (g *GMap[K, D]) GetHash() uint64 {
	hash := uint64(0)

	g.lock.RLock()

	for _, value := range g.contents {
		hash += value.Vector
	}

	g.lock.RUnlock()
	return hash
}

// Prune: prune all stale entries
func (g *GMap[K, D]) Prune() {
	stale := g.clock.getStale()
	g.lock.Lock()

	for _, outlier := range stale {
		delete(g.contents, outlier)
	}

	g.lock.Unlock()
}

func NewGMap[K cmp.Ordered, D any](clock *VectorClock[K]) *GMap[K, D] {
	return &GMap[K, D]{
		contents: make(map[uint64]Bucket[D]),
		clock:    clock,
	}
}
