package sync

import (
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

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
