package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	ctrlserver "github.com/tim-beatham/smegmesh/pkg/ctrlserver"
	"github.com/tim-beatham/smegmesh/pkg/ipc"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/robin"
	"github.com/tim-beatham/smegmesh/pkg/sync"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func main() {
	if len(os.Args) != 2 {
		logging.Log.WriteErrorf("Did not provide configuration")
		return
	}

	conf, err := conf.ParseDaemonConfiguration(os.Args[1])
	if err != nil {
		logging.Log.WriteErrorf("Could not parse configuration: %s", err.Error())
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

	if err != nil {
		panic(err)
	}

	syncProvider.MeshManager = ctrlServer.MeshManager

	robinIpcParams := robin.RobinIpcParams{
		CtrlServer: ctrlServer,
	}

	robinRpc.Server = ctrlServer
	robinIpc = robin.NewRobinIpc(robinIpcParams)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return
	}

	logging.Log.WriteInfof("running ipc handler")
	go ipc.RunIpcHandler(&robinIpc)

	closeResources := func() {
		logging.Log.WriteInfof("closing resources")
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
