package main

import (
	"fmt"

	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/api"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/ipc"
	wg "github.com/tim-beatham/wgmesh/pkg/wg"
)

func main() {
	wgClient, err := wg.CreateClient("wgmesh")

	if err != nil {
		fmt.Println(err)
		return
	}

	ctrlServer := ctrlserver.NewCtrlServer("0.0.0.0", 21910, wgClient)
	ipc.RunIpcHandler(ctrlServer)
	r := api.RunAPI(ctrlServer)
	r.Run(ctrlServer.GetEndpoint())
}
