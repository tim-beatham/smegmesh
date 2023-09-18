package main

import (
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/api"
)

func main() {
	ctrlServer := ctrlserver.NewCtrlServer("0.0.0.0", 21910)
	r := api.RunAPI(ctrlServer)
	r.Run(ctrlServer.GetEndpoint())
}
