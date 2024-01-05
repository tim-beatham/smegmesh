// automerge: automerge is a CRDT library. Defines a CRDT
// datastore and methods to resolve conflicts
package automerge

import (
	"github.com/automerge/automerge-go"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
)

// AutomergeSync: defines a synchroniser to bi-directionally synchronise the
// two states
type AutomergeSync struct {
	// state: the automerge sync state to use
	state *automerge.SyncState
	// manager: the corresponding data store that we are merging
	manager *CrdtMeshManager
}

// GenerateMessage: geenrate a new automerge message to synchronise
// returns a byte of the message and a boolean of whether or not there
// are more messages in the sequence
func (a *AutomergeSync) GenerateMessage() ([]byte, bool) {
	msg, valid := a.state.GenerateMessage()

	if !valid {
		return nil, false
	}

	return msg.Bytes(), true
}

// RecvMessage: receive an automerge message to merge in the datastore
// returns an error if unsuccessful
func (a *AutomergeSync) RecvMessage(msg []byte) error {
	_, err := a.state.ReceiveMessage(msg)

	if err != nil {
		return err
	}

	return nil
}

// Complete: complete the synchronisation process
func (a *AutomergeSync) Complete() {
	logging.Log.WriteInfof("sync completed")
	a.manager.SaveChanges()
}

// NewAutomergeSync: instantiates a new automerge syncer
func NewAutomergeSync(manager *CrdtMeshManager) *AutomergeSync {
	return &AutomergeSync{
		state:   automerge.NewSyncState(manager.doc),
		manager: manager,
	}
}
