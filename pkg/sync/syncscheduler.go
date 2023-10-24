package sync

import (
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

// SyncScheduler: Loops through all nodes in the mesh and runs a schedule to
// sync each event
type SyncScheduler interface {
	Run() error
	Stop() error
}

// SyncSchedulerImpl scheduler for sync scheduling
type SyncSchedulerImpl struct {
	syncRate int
	quit     chan struct{}
	server   *ctrlserver.MeshCtrlServer
	syncer   Syncer
}

// Run implements SyncScheduler.
func (s *SyncSchedulerImpl) Run() error {
	ticker := time.NewTicker(time.Duration(s.syncRate) * time.Second)

	quit := make(chan struct{})
	s.quit = quit

	for {
		select {
		case <-ticker.C:
			err := s.syncer.SyncMeshes()

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
			}
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

func NewSyncScheduler(s *ctrlserver.MeshCtrlServer, syncRequester SyncRequester, syncRate int) SyncScheduler {
	syncer := NewSyncer(s.MeshManager, syncRequester)
	return &SyncSchedulerImpl{server: s, syncRate: syncRate, syncer: syncer}
}
