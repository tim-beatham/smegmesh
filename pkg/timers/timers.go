package timer

import (
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

func NewTimestampScheduler(ctrlServer *ctrlserver.MeshCtrlServer) lib.Timer {
	timerFunc := func() error {
		return ctrlServer.MeshManager.UpdateTimeStamp()
	}

	return *lib.NewTimer(timerFunc, ctrlServer.Conf.KeepAliveTime)
}

func NewRouteScheduler(ctrlServer *ctrlserver.MeshCtrlServer) lib.Timer {
	timerFunc := func() error {
		return ctrlServer.MeshManager.GetRouteManager().UpdateRoutes()
	}

	return *lib.NewTimer(timerFunc, 10)
}
