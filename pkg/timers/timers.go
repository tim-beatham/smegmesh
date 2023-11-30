package timer

import (
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

func NewTimestampScheduler(ctrlServer *ctrlserver.MeshCtrlServer) lib.Timer {
	timerFunc := func() error {
		logging.Log.WriteInfof("Updated Timestamp")
		return ctrlServer.MeshManager.UpdateTimeStamp()
	}

	return *lib.NewTimer(timerFunc, ctrlServer.Conf.KeepAliveTime)
}
