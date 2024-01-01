package automerge

import (
	"github.com/automerge/automerge-go"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
)

type AutomergeSync struct {
	state   *automerge.SyncState
	manager *CrdtMeshManager
}

func (a *AutomergeSync) GenerateMessage() ([]byte, bool) {
	msg, valid := a.state.GenerateMessage()

	if !valid {
		return nil, false
	}

	return msg.Bytes(), true
}

func (a *AutomergeSync) RecvMessage(msg []byte) error {
	_, err := a.state.ReceiveMessage(msg)

	if err != nil {
		return err
	}

	return nil
}

func (a *AutomergeSync) Complete() {
	logging.Log.WriteInfof("Sync Completed")
	a.manager.SaveChanges()
}

func NewAutomergeSync(manager *CrdtMeshManager) *AutomergeSync {
	return &AutomergeSync{
		state:   automerge.NewSyncState(manager.doc),
		manager: manager,
	}
}
