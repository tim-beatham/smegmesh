package sync

import (
	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
	"github.com/tim-beatham/smegmesh/pkg/lib"
)

// Run implements SyncScheduler.
func syncFunction(syncer Syncer) lib.TimerFunc {
	return func() error {
		syncer.SyncMeshes()
		return nil
	}
}

func NewSyncScheduler(s *ctrlserver.MeshCtrlServer, syncRequester SyncRequester, syncer Syncer) *lib.Timer {
	return lib.NewTimer(syncFunction(syncer), s.Conf.SyncTime)
}
