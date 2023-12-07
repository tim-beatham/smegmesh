// crdt is a golang implementation of a crdt
package crdt

import (
	"cmp"
	"sync"
)

type Bucket[D any] struct {
	Vector     uint64
	Contents   D
	Gravestone bool
}

// GMap is a set that can only grow in size
type GMap[K cmp.Ordered, D any] struct {
	lock     sync.RWMutex
	contents map[K]Bucket[D]
	clock    *VectorClock[K]
}

func (g *GMap[K, D]) Put(key K, value D) {
	g.lock.Lock()

	clock := g.clock.IncrementClock()

	g.contents[key] = Bucket[D]{
		Vector:   clock,
		Contents: value,
	}

	g.lock.Unlock()
}

func (g *GMap[K, D]) Contains(key K) bool {
	g.lock.RLock()

	_, ok := g.contents[key]

	g.lock.RUnlock()

	return ok
}

func (g *GMap[K, D]) put(key K, b Bucket[D]) {
	g.lock.Lock()

	if g.contents[key].Vector < b.Vector {
		g.contents[key] = b
	}

	g.lock.Unlock()
}

func (g *GMap[K, D]) get(key K) Bucket[D] {
	g.lock.RLock()
	bucket := g.contents[key]
	g.lock.RUnlock()

	return bucket
}

func (g *GMap[K, D]) Get(key K) D {
	return g.get(key).Contents
}

func (g *GMap[K, D]) Mark(key K) {
	g.lock.Lock()
	bucket := g.contents[key]
	bucket.Gravestone = true
	g.contents[key] = bucket
	g.lock.Unlock()
}

// IsMarked: returns true if the node is marked
func (g *GMap[K, D]) IsMarked(key K) bool {
	marked := false

	g.lock.RLock()

	bucket, ok := g.contents[key]

	if ok {
		marked = bucket.Gravestone
	}

	g.lock.RUnlock()

	return marked
}

func (g *GMap[K, D]) Keys() []K {
	g.lock.RLock()

	contents := make([]K, len(g.contents))
	index := 0

	for key := range g.contents {
		contents[index] = key
		index++
	}

	g.lock.RUnlock()
	return contents
}

func (g *GMap[K, D]) Save() map[K]Bucket[D] {
	buckets := make(map[K]Bucket[D])
	g.lock.RLock()

	for key, value := range g.contents {
		buckets[key] = value
	}

	g.lock.RUnlock()
	return buckets
}

func (g *GMap[K, D]) SaveWithKeys(keys []K) map[K]Bucket[D] {
	buckets := make(map[K]Bucket[D])
	g.lock.RLock()

	for _, key := range keys {
		buckets[key] = g.contents[key]
	}

	g.lock.RUnlock()
	return buckets
}

func (g *GMap[K, D]) GetClock() map[K]uint64 {
	clock := make(map[K]uint64)
	g.lock.RLock()

	for key, bucket := range g.contents {
		clock[key] = bucket.Vector
	}

	g.lock.RUnlock()
	return clock
}

func (g *GMap[K, D]) GetHash() uint64 {
	hash := uint64(0)

	g.lock.RLock()

	for _, value := range g.contents {
		hash += value.Vector
	}

	g.lock.RUnlock()
	return hash
}

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
		contents: make(map[K]Bucket[D]),
		clock:    clock,
	}
}
