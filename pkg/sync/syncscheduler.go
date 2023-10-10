package sync

import (
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
)

// SyncScheduler: Loops through all nodes in the mesh and runs a schedule to
// sync each event
type SyncScheduler interface {
	Run() error
	Stop() error
}

type SyncSchedulerImpl struct {
	quit   chan struct{}
	server *ctrlserver.MeshCtrlServer
}

// Run implements SyncScheduler.
func (s *SyncSchedulerImpl) Run() error {
	ticker := time.NewTicker(time.Second)

	quit := make(chan struct{})
	s.quit = quit

	for {
		select {
		case <-ticker.C:
			break
		case <-quit:
			break
		}
	}
}

// Stop implements SyncScheduler.
func (s *SyncSchedulerImpl) Stop() error {
	close(s.quit)
	return nil
}

func NewSyncScheduler(s *ctrlserver.MeshCtrlServer) SyncScheduler {
	return &SyncSchedulerImpl{server: s}
}
