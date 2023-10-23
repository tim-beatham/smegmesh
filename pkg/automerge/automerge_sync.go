package crdt

import (
	"github.com/automerge/automerge-go"
)

type AutomergeSync struct {
	state *automerge.SyncState
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

func NewAutomergeSync(manager *CrdtNodeManager) *AutomergeSync {
	return &AutomergeSync{state: automerge.NewSyncState(manager.doc)}
}
