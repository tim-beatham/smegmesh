package crdt

import (
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

type TwoPhaseMap[K comparable, D any] struct {
	addMap    *GMap[K, D]
	removeMap *GMap[K, bool]
	Clock     *VectorClock[K]
	processId K
}

type TwoPhaseMapSnapshot[K comparable, D any] struct {
	Add    map[K]Bucket[D]
	Remove map[K]Bucket[bool]
}

// Contains checks whether the value exists in the map
func (m *TwoPhaseMap[K, D]) Contains(key K) bool {
	if !m.addMap.Contains(key) {
		return false
	}

	addValue := m.addMap.get(key)

	if !m.removeMap.Contains(key) {
		return true
	}

	removeValue := m.removeMap.get(key)

	return addValue.Vector >= removeValue.Vector
}

func (m *TwoPhaseMap[K, D]) Get(key K) D {
	var result D

	if !m.Contains(key) {
		return result
	}

	return m.addMap.Get(key)
}

// Put places the key K in the map
func (m *TwoPhaseMap[K, D]) Put(key K, data D) {
	msgSequence := m.Clock.IncrementClock()
	m.Clock.Put(key, msgSequence)
	m.addMap.Put(key, data)
}

func (m *TwoPhaseMap[K, D]) Mark(key K) {
	m.addMap.Mark(key)
}

// Remove removes the value from the map
func (m *TwoPhaseMap[K, D]) Remove(key K) {
	m.removeMap.Put(key, true)
}

func (m *TwoPhaseMap[K, D]) Keys() []K {
	keys := make([]K, 0)

	addKeys := m.addMap.Keys()

	for _, key := range addKeys {
		if !m.Contains(key) {
			continue
		}

		keys = append(keys, key)
	}

	return keys
}

func (m *TwoPhaseMap[K, D]) AsMap() map[K]D {
	theMap := make(map[K]D)

	keys := m.Keys()

	for _, key := range keys {
		theMap[key] = m.Get(key)
	}

	return theMap
}

func (m *TwoPhaseMap[K, D]) Snapshot() *TwoPhaseMapSnapshot[K, D] {
	return &TwoPhaseMapSnapshot[K, D]{
		Add:    m.addMap.Save(),
		Remove: m.removeMap.Save(),
	}
}

func (m *TwoPhaseMap[K, D]) SnapShotFromState(state *TwoPhaseMapState[K]) *TwoPhaseMapSnapshot[K, D] {
	addKeys := lib.MapKeys(state.AddContents)
	removeKeys := lib.MapKeys(state.RemoveContents)

	return &TwoPhaseMapSnapshot[K, D]{
		Add:    m.addMap.SaveWithKeys(addKeys),
		Remove: m.removeMap.SaveWithKeys(removeKeys),
	}
}

type TwoPhaseMapState[K comparable] struct {
	Vectors        map[K]uint64
	AddContents    map[K]uint64
	RemoveContents map[K]uint64
}

func (m *TwoPhaseMap[K, D]) IsMarked(key K) bool {
	return m.addMap.IsMarked(key)
}

// GetHash: Get the hash of the current state of the map
// Sums the current values of the vectors. Provides good approximation
// of increasing numbers
func (m *TwoPhaseMap[K, D]) GetHash() uint64 {
	return m.addMap.GetHash() + m.removeMap.GetHash()
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

func (m *TwoPhaseMap[K, D]) UpdateVector(state *TwoPhaseMapState[K]) {
	for key, value := range state.Vectors {
		m.Clock.Put(key, value)
	}
}

func (m *TwoPhaseMapState[K]) Difference(state *TwoPhaseMapState[K]) *TwoPhaseMapState[K] {
	mapState := &TwoPhaseMapState[K]{
		AddContents:    make(map[K]uint64),
		RemoveContents: make(map[K]uint64),
	}

	for key, value := range state.AddContents {
		otherValue, ok := m.AddContents[key]

		if !ok || otherValue < value {
			mapState.AddContents[key] = value
		}
	}

	for key, value := range state.AddContents {
		otherValue, ok := m.RemoveContents[key]

		if !ok || otherValue < value {
			mapState.RemoveContents[key] = value
		}
	}

	return mapState
}

func (m *TwoPhaseMap[K, D]) Merge(snapshot TwoPhaseMapSnapshot[K, D]) {
	for key, value := range snapshot.Add {
		// Gravestone is local only to that node.
		// Discover ourselves if the node is alive
		m.addMap.put(key, value)
		m.Clock.Put(key, value.Vector)
	}

	for key, value := range snapshot.Remove {
		m.removeMap.put(key, value)
		m.Clock.Put(key, value.Vector)
	}
}

func (m *TwoPhaseMap[K, D]) Prune() {
	m.addMap.Prune()
	m.removeMap.Prune()
	m.Clock.Prune()
}

// NewTwoPhaseMap: create a new two phase map. Consists of two maps
// a grow map and a remove map. If both timestamps equal then favour keeping
// it in the map
func NewTwoPhaseMap[K comparable, D any](processId K) *TwoPhaseMap[K, D] {
	m := TwoPhaseMap[K, D]{
		processId: processId,
		Clock:     NewVectorClock(processId),
	}

	m.addMap = NewGMap[K, D](m.Clock)
	m.removeMap = NewGMap[K, bool](m.Clock)
	return &m
}
