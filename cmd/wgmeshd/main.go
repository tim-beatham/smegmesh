package main

import (
	"log"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/middleware"
	"github.com/tim-beatham/wgmesh/pkg/robin"
	"github.com/tim-beatham/wgmesh/pkg/sync"
	wg "github.com/tim-beatham/wgmesh/pkg/wg"
)

func main() {
	conf, err := conf.ParseConfiguration("./configuration.yaml")
	if err != nil {
		log.Fatalln("Could not parse configuration")
	}

	wgClient, err := wg.CreateClient(conf.IfName, conf.WgPort)

	var robinRpc robin.RobinRpc
	var robinIpc robin.RobinIpc
	var authProvider middleware.AuthRpcProvider
	var syncProvider sync.SyncServiceImpl

	ctrlServerParams := ctrlserver.NewCtrlServerParams{
		WgClient:     wgClient,
		Conf:         conf,
		AuthProvider: &authProvider,
		CtrlProvider: &robinRpc,
		SyncProvider: &syncProvider,
	}

	ctrlServer, err := ctrlserver.NewCtrlServer(&ctrlServerParams)
	syncProvider.Server = ctrlServer
	syncRequester := sync.NewSyncRequester(ctrlServer)
	syncScheduler := sync.NewSyncScheduler(ctrlServer, syncRequester, 2)

	robinIpcParams := robin.RobinIpcParams{
		CtrlServer: ctrlServer,
		Allocator:  &ip.ULABuilder{},
	}

	robinRpc.Server = ctrlServer
	robinIpc = robin.NewRobinIpc(robinIpcParams)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return
	}

	log.Println("Running IPC Handler")

	go ipc.RunIpcHandler(&robinIpc)
	go syncScheduler.Run()

	err = ctrlServer.ConnectionServer.Listen()

	if err != nil {
		logging.Log.WriteErrorf(err.Error())

		return
	}

	defer syncScheduler.Stop()
	defer ctrlServer.Close()
	defer wgClient.Close()
}
