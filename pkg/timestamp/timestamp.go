package timestamp

import (
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

type TimestampScheduler interface {
	Run() error
	Stop() error
}

type TimeStampSchedulerImpl struct {
	meshManager mesh.MeshManager
	updateRate  int
	quit        chan struct{}
}

func (s *TimeStampSchedulerImpl) Run() error {
	ticker := time.NewTicker(time.Duration(s.updateRate) * time.Second)

	s.quit = make(chan struct{})

	for {
		select {
		case <-ticker.C:
			err := s.meshManager.UpdateTimeStamp()

			if err != nil {
				logging.Log.WriteErrorf("Update Timestamp Error: %s", err.Error())
			}
		case <-s.quit:
			break
		}
	}
}

func NewTimestampScheduler(ctrlServer *ctrlserver.MeshCtrlServer) TimestampScheduler {
	return &TimeStampSchedulerImpl{
		meshManager: ctrlServer.MeshManager,
		updateRate:  ctrlServer.Conf.KeepAliveRate,
	}
}

func (s *TimeStampSchedulerImpl) Stop() error {
	close(s.quit)
	return nil
}
