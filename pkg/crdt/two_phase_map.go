package crdt

import (
	"sync"

	"github.com/tim-beatham/wgmesh/pkg/lib"
)

type TwoPhaseMap[K comparable, D any] struct {
	addMap    *GMap[K, D]
	removeMap *GMap[K, bool]
	vectors   map[K]uint64
	processId K
	lock      sync.RWMutex
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
	msgSequence := m.incrementClock()

	m.lock.Lock()

	if _, ok := m.vectors[key]; !ok {
		m.vectors[key] = msgSequence
	}

	m.lock.Unlock()
	m.addMap.Put(key, data)
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
	AddContents    map[K]uint64
	RemoveContents map[K]uint64
}

func (m *TwoPhaseMap[K, D]) incrementClock() uint64 {
	maxClock := uint64(0)
	m.lock.Lock()

	for _, value := range m.vectors {
		maxClock = max(maxClock, value)
	}

	m.vectors[m.processId] = maxClock + 1
	m.lock.Unlock()
	return maxClock
}

func (m *TwoPhaseMap[K, D]) GetClock() uint64 {
	maxClock := uint64(0)
	m.lock.RLock()

	for _, value := range m.vectors {
		maxClock = max(maxClock, value)
	}

	m.lock.RUnlock()
	return maxClock
}

// GetState: get the current vector clock of the add and remove
// map
func (m *TwoPhaseMap[K, D]) GenerateMessage() *TwoPhaseMapState[K] {
	addContents := m.addMap.GetClock()
	removeContents := m.removeMap.GetClock()

	return &TwoPhaseMapState[K]{
		AddContents:    addContents,
		RemoveContents: removeContents,
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
	m.lock.Lock()

	for key, value := range snapshot.Add {
		m.addMap.put(key, value)
		m.vectors[key] = max(value.Vector, m.vectors[key])
	}

	for key, value := range snapshot.Remove {
		m.removeMap.put(key, value)
		m.vectors[key] = max(value.Vector, m.vectors[key])
	}

	m.lock.Unlock()
}

// NewTwoPhaseMap: create a new two phase map. Consists of two maps
// a grow map and a remove map. If both timestamps equal then favour keeping
// it in the map
func NewTwoPhaseMap[K comparable, D any](processId K) *TwoPhaseMap[K, D] {
	m := TwoPhaseMap[K, D]{
		vectors:   make(map[K]uint64),
		processId: processId,
	}

	m.addMap = NewGMap[K, D](m.incrementClock)
	m.removeMap = NewGMap[K, bool](m.incrementClock)
	return &m
}
