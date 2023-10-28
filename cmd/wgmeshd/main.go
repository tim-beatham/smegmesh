package main

import (
	"log"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/middleware"
	"github.com/tim-beatham/wgmesh/pkg/robin"
	"github.com/tim-beatham/wgmesh/pkg/sync"
	"github.com/tim-beatham/wgmesh/pkg/timestamp"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func main() {
	conf, err := conf.ParseConfiguration("./configuration.yaml")
	if err != nil {
		log.Fatalln("Could not parse configuration")
	}

	client, err := wgctrl.New()

	if err != nil {
		logging.Log.WriteErrorf("Failed to create wgctrl client")
		return
	}

	var robinRpc robin.WgRpc
	var robinIpc robin.IpcHandler
	var authProvider middleware.AuthRpcProvider
	var syncProvider sync.SyncServiceImpl

	ctrlServerParams := ctrlserver.NewCtrlServerParams{
		Conf:         conf,
		AuthProvider: &authProvider,
		CtrlProvider: &robinRpc,
		SyncProvider: &syncProvider,
		Client:       client,
	}

	ctrlServer, err := ctrlserver.NewCtrlServer(&ctrlServerParams)
	syncProvider.Server = ctrlServer
	syncRequester := sync.NewSyncRequester(ctrlServer)
	syncScheduler := sync.NewSyncScheduler(ctrlServer, syncRequester, 2)
	timestampScheduler := timestamp.NewTimestampScheduler(ctrlServer, 60)

	robinIpcParams := robin.RobinIpcParams{
		CtrlServer: ctrlServer,
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
	go timestampScheduler.Run()

	err = ctrlServer.ConnectionServer.Listen()

	if err != nil {
		logging.Log.WriteErrorf(err.Error())

		return
	}

	defer syncScheduler.Stop()
	defer timestampScheduler.Stop()
	defer ctrlServer.Close()
	defer client.Close()
}
