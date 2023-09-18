package main

import (
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/api"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/ipc"
)

func main() {
	ctrlServer := ctrlserver.NewCtrlServer("0.0.0.0", 21910)
	r := api.RunAPI(ctrlServer)
	ipc.RunIpcHandler()
	r.Run(ctrlServer.GetEndpoint())
}
