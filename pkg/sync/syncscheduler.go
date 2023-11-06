package sync

import (
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

// SyncScheduler: Loops through all nodes in the mesh and runs a schedule to
// sync each event
type SyncScheduler interface {
	Run() error
	Stop() error
}

// SyncSchedulerImpl scheduler for sync scheduling
type SyncSchedulerImpl struct {
	quit   chan struct{}
	server *ctrlserver.MeshCtrlServer
	syncer Syncer
}

// Run implements SyncScheduler.
func syncFunction(syncer Syncer) lib.TimerFunc {
	return func() error {
		return syncer.SyncMeshes()
	}
}

func NewSyncScheduler(s *ctrlserver.MeshCtrlServer, syncRequester SyncRequester) *lib.Timer {
	syncer := NewSyncer(s.MeshManager, s.Conf, syncRequester)
	return lib.NewTimer(syncFunction(syncer), int(s.Conf.SyncRate))
}
