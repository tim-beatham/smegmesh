package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/robin"
	"github.com/tim-beatham/wgmesh/pkg/sync"
	timer "github.com/tim-beatham/wgmesh/pkg/timers"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func main() {
	if len(os.Args) != 2 {
		logging.Log.WriteErrorf("Need to provide configuration.yaml")
		return
	}

	conf, err := conf.ParseConfiguration(os.Args[1])
	if err != nil {
		logging.Log.WriteInfof("Could not parse configuration")
		return
	}

	client, err := wgctrl.New()

	if err != nil {
		logging.Log.WriteErrorf("Failed to create wgctrl client")
		return
	}

	if conf.Profile {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	var robinRpc robin.WgRpc
	var robinIpc robin.IpcHandler
	var syncProvider sync.SyncServiceImpl

	ctrlServerParams := ctrlserver.NewCtrlServerParams{
		Conf:         conf,
		CtrlProvider: &robinRpc,
		SyncProvider: &syncProvider,
		Client:       client,
	}

	ctrlServer, err := ctrlserver.NewCtrlServer(&ctrlServerParams)
	syncProvider.Server = ctrlServer
	syncRequester := sync.NewSyncRequester(ctrlServer)
	syncScheduler := sync.NewSyncScheduler(ctrlServer, syncRequester)
	timestampScheduler := timer.NewTimestampScheduler(ctrlServer)
	pruneScheduler := mesh.NewPruner(ctrlServer.MeshManager, *conf)
	routeScheduler := timer.NewRouteScheduler(ctrlServer)

	robinIpcParams := robin.RobinIpcParams{
		CtrlServer: ctrlServer,
	}

	robinRpc.Server = ctrlServer
	robinIpc = robin.NewRobinIpc(robinIpcParams)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return
	}

	logging.Log.WriteInfof("Running IPC Handler")

	go ipc.RunIpcHandler(&robinIpc)
	go syncScheduler.Run()
	go timestampScheduler.Run()
	go pruneScheduler.Run()
	go routeScheduler.Run()

	closeResources := func() {
		logging.Log.WriteInfof("Closing resources")
		syncScheduler.Stop()
		timestampScheduler.Stop()
		ctrlServer.Close()
		client.Close()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			closeResources()
			os.Exit(0)
		}
	}()

	err = ctrlServer.ConnectionServer.Listen()

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return
	}

	go closeResources()
}
