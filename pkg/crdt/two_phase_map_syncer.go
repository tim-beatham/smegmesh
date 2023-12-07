package crdt

import (
	"bytes"
	"encoding/gob"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

type SyncState int

const (
	HASH SyncState = iota
	PREPARE
	PRESENT
	EXCHANGE
	MERGE
	FINISHED
)

// TwoPhaseSyncer is a type to sync a TwoPhase data store
type TwoPhaseSyncer struct {
	manager            *TwoPhaseStoreMeshManager
	generateMessageFSM SyncFSM
	state              SyncState
	mapState           *TwoPhaseMapState[string]
	peerMsg            []byte
}

type TwoPhaseHash struct {
	Hash uint64
}

type SyncFSM map[SyncState]func(*TwoPhaseSyncer) ([]byte, bool)

func hash(syncer *TwoPhaseSyncer) ([]byte, bool) {
	hash := TwoPhaseHash{
		Hash: syncer.manager.store.Clock.GetHash(),
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(hash)

	if err != nil {
		logging.Log.WriteInfof(err.Error())
	}

	syncer.IncrementState()
	return buffer.Bytes(), true
}

func prepare(syncer *TwoPhaseSyncer) ([]byte, bool) {
	var recvBuffer = bytes.NewBuffer(syncer.peerMsg)
	dec := gob.NewDecoder(recvBuffer)

	var hash TwoPhaseHash
	err := dec.Decode(&hash)

	if err != nil {
		logging.Log.WriteInfof(err.Error())
	}

	// If vector clocks are equal then no need to merge state
	// Helps to reduce bandwidth by detecting early
	if hash.Hash == syncer.manager.store.Clock.GetHash() {
		return nil, false
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err = enc.Encode(*syncer.mapState)

	if err != nil {
		logging.Log.WriteInfof(err.Error())
	}

	syncer.IncrementState()
	return buffer.Bytes(), true
}

func present(syncer *TwoPhaseSyncer) ([]byte, bool) {
	if syncer.peerMsg == nil {
		panic("peer msg is nil")
	}

	var recvBuffer = bytes.NewBuffer(syncer.peerMsg)
	dec := gob.NewDecoder(recvBuffer)

	var mapState TwoPhaseMapState[string]
	err := dec.Decode(&mapState)

	if err != nil {
		logging.Log.WriteInfof(err.Error())
	}

	difference := syncer.mapState.Difference(&mapState)

	var sendBuffer bytes.Buffer
	enc := gob.NewEncoder(&sendBuffer)
	enc.Encode(*difference)

	syncer.IncrementState()
	return sendBuffer.Bytes(), true
}

func exchange(syncer *TwoPhaseSyncer) ([]byte, bool) {
	if syncer.peerMsg == nil {
		panic("peer msg is nil")
	}

	var recvBuffer = bytes.NewBuffer(syncer.peerMsg)
	dec := gob.NewDecoder(recvBuffer)

	var mapState TwoPhaseMapState[string]
	dec.Decode(&mapState)

	snapshot := syncer.manager.store.SnapShotFromState(&mapState)

	var sendBuffer bytes.Buffer
	enc := gob.NewEncoder(&sendBuffer)
	enc.Encode(*snapshot)

	syncer.IncrementState()
	return sendBuffer.Bytes(), true
}

func merge(syncer *TwoPhaseSyncer) ([]byte, bool) {
	if syncer.peerMsg == nil {
		panic("peer msg is nil")
	}

	var recvBuffer = bytes.NewBuffer(syncer.peerMsg)
	dec := gob.NewDecoder(recvBuffer)

	var snapshot TwoPhaseMapSnapshot[string, MeshNode]
	dec.Decode(&snapshot)

	syncer.manager.store.Merge(snapshot)
	return nil, false
}

func (t *TwoPhaseSyncer) IncrementState() {
	t.state = min(t.state+1, FINISHED)
}

func (t *TwoPhaseSyncer) GenerateMessage() ([]byte, bool) {
	fsmFunc, ok := t.generateMessageFSM[t.state]

	if !ok {
		panic("state not handled")
	}

	return fsmFunc(t)
}

func (t *TwoPhaseSyncer) RecvMessage(msg []byte) error {
	t.peerMsg = msg
	return nil
}

func (t *TwoPhaseSyncer) Complete() {
	logging.Log.WriteInfof("SYNC COMPLETED")
	if t.state == FINISHED || t.state == MERGE {
		t.manager.store.Clock.IncrementClock()
	}
}

func NewTwoPhaseSyncer(manager *TwoPhaseStoreMeshManager) *TwoPhaseSyncer {
	var generateMessageFsm SyncFSM = SyncFSM{
		HASH:     hash,
		PREPARE:  prepare,
		PRESENT:  present,
		EXCHANGE: exchange,
		MERGE:    merge,
	}

	return &TwoPhaseSyncer{
		manager:            manager,
		state:              HASH,
		mapState:           manager.store.GenerateMessage(),
		generateMessageFSM: generateMessageFsm,
	}
}
