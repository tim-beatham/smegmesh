package crdt

import (
	"cmp"

	"github.com/tim-beatham/smegmesh/pkg/lib"
)

// TwoPhaseMap: comprises of two grow-only maps
type TwoPhaseMap[K cmp.Ordered, D any] struct {
	addMap    *GMap[K, D]
	removeMap *GMap[K, bool]
	Clock     *VectorClock[K]
	processId K
}

type TwoPhaseMapSnapshot[K cmp.Ordered, D any] struct {
	Add    map[uint64]Bucket[D]
	Remove map[uint64]Bucket[bool]
}

// Contains checks whether the value exists in the map
func (m *TwoPhaseMap[K, D]) Contains(key K) bool {
	return m.contains(m.Clock.hashFunc(key))
}

// contains: checks whether the key exists in the map
func (m *TwoPhaseMap[K, D]) contains(key uint64) bool {
	if !m.addMap.contains(key) {
		return false
	}

	addValue := m.addMap.get(key)

	if !m.removeMap.contains(key) {
		return true
	}

	removeValue := m.removeMap.get(key)

	return addValue.Vector >= removeValue.Vector
}

// Get: get the value corresponding with the given key
func (m *TwoPhaseMap[K, D]) Get(key K) D {
	var result D

	if !m.Contains(key) {
		return result
	}

	return m.addMap.Get(key)
}

func (m *TwoPhaseMap[K, D]) get(key uint64) D {
	var result D

	if !m.contains(key) {
		return result
	}

	return m.addMap.get(key).Contents
}

// Put: places the key K in the map with the associated data D
func (m *TwoPhaseMap[K, D]) Put(key K, data D) {
	msgSequence := m.Clock.IncrementClock()
	m.Clock.Put(key, msgSequence)
	m.addMap.Put(key, data)
}

// Mark: marks the status of the node as undetermiend
func (m *TwoPhaseMap[K, D]) Mark(key K) {
	m.addMap.Mark(key)
}

// Remove: removes the value from the map
func (m *TwoPhaseMap[K, D]) Remove(key K) {
	m.removeMap.Put(key, true)
}

func (m *TwoPhaseMap[K, D]) keys() []uint64 {
	keys := make([]uint64, 0)

	addKeys := m.addMap.Keys()

	for _, key := range addKeys {
		if !m.contains(key) {
			continue
		}

		keys = append(keys, key)
	}

	return keys
}

// AsList: convert the map to a list
func (m *TwoPhaseMap[K, D]) AsList() []D {
	theList := make([]D, 0)

	keys := m.keys()

	for _, key := range keys {
		theList = append(theList, m.get(key))
	}

	return theList
}

// Snapshot: convert the map into an immutable snapshot.
// contains the contents of the add and remove map
func (m *TwoPhaseMap[K, D]) Snapshot() *TwoPhaseMapSnapshot[K, D] {
	return &TwoPhaseMapSnapshot[K, D]{
		Add:    m.addMap.Save(),
		Remove: m.removeMap.Save(),
	}
}

// SnapshotFromState: create a snapshot of the intersection of values provided
// in the given state
func (m *TwoPhaseMap[K, D]) SnapShotFromState(state *TwoPhaseMapState[K]) *TwoPhaseMapSnapshot[K, D] {
	addKeys := lib.MapKeys(state.AddContents)
	removeKeys := lib.MapKeys(state.RemoveContents)

	return &TwoPhaseMapSnapshot[K, D]{
		Add:    m.addMap.SaveWithKeys(addKeys),
		Remove: m.removeMap.SaveWithKeys(removeKeys),
	}
}

// TwoPhaseMapState: encapsulates the state of the map
// without specifying the data that is stored
type TwoPhaseMapState[K cmp.Ordered] struct {
	// Vectors: the vector ID of each process
	Vectors        map[uint64]uint64
	// AddContents: the contents of the add map
	AddContents    map[uint64]uint64
	// RemoveContents: the contents of the remove map
	RemoveContents map[uint64]uint64
}

// IsMarked: returns true if the given value is marked in an undetermined state
func (m *TwoPhaseMap[K, D]) IsMarked(key K) bool {
	return m.addMap.IsMarked(key)
}

// GetHash: Get the hash of the current state of the map
// Sums the current values of the vectors. Provides good approximation
// of increasing numbers
func (m *TwoPhaseMap[K, D]) GetHash() uint64 {
	return (m.addMap.GetHash() + 1) * (m.removeMap.GetHash() + 1)
}

// GetState: get the current vector clock of the add and remove
// map
func (m *TwoPhaseMap[K, D]) GenerateMessage() *TwoPhaseMapState[K] {
	addContents := m.addMap.GetClock()
	removeContents := m.removeMap.GetClock()

	return &TwoPhaseMapState[K]{
		Vectors:        m.Clock.GetClock(),
		AddContents:    addContents,
		RemoveContents: removeContents,
	}
}

// Difference: compute the set difference between the two states.
// highestStale represents the highest vector clock that has been marked as stale
func (m *TwoPhaseMapState[K]) Difference(highestStale uint64, state *TwoPhaseMapState[K]) *TwoPhaseMapState[K] {
	mapState := &TwoPhaseMapState[K]{
		AddContents:    make(map[uint64]uint64),
		RemoveContents: make(map[uint64]uint64),
	}

	for key, value := range state.AddContents {
		otherValue, ok := m.AddContents[key]

		if value > highestStale && (!ok || otherValue < value) {
			mapState.AddContents[key] = value
		}
	}

	for key, value := range state.RemoveContents {
		otherValue, ok := m.RemoveContents[key]

		if value > highestStale && (!ok || otherValue < value) {
			mapState.RemoveContents[key] = value
		}
	}

	return mapState
}

// Merge: merge a snapshot into the map
func (m *TwoPhaseMap[K, D]) Merge(snapshot TwoPhaseMapSnapshot[K, D]) {
	for key, value := range snapshot.Add {
		// Gravestone is local only to that node.
		// Discover ourselves if the node is alive
		m.addMap.put(key, value)
		m.Clock.put(key, value.Vector)
	}

	for key, value := range snapshot.Remove {
		m.removeMap.put(key, value)
		m.Clock.put(key, value.Vector)
	}
}

// Prune: garbage collect all stale entries in the map
func (m *TwoPhaseMap[K, D]) Prune() {
	m.addMap.Prune()
	m.removeMap.Prune()
	m.Clock.Prune()
}

// NewTwoPhaseMap: create a new two phase map. Consists of two maps
// a grow map and a remove map. If both timestamps equal then favour keeping
// it in the map
func NewTwoPhaseMap[K cmp.Ordered, D any](processId K, hashKey func(K) uint64, staleTime uint64) *TwoPhaseMap[K, D] {
	m := TwoPhaseMap[K, D]{
		processId: processId,
		Clock:     NewVectorClock(processId, hashKey, staleTime),
	}

	m.addMap = NewGMap[K, D](m.Clock)
	m.removeMap = NewGMap[K, bool](m.Clock)
	return &m
}
