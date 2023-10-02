package main

import (
	"log"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/middleware"
	"github.com/tim-beatham/wgmesh/pkg/robin"
	wg "github.com/tim-beatham/wgmesh/pkg/wg"
)

func main() {
	conf, err := conf.ParseConfiguration("./configuration.yaml")
	if err != nil {
		log.Fatalln("Could not parse configuration")
	}

	wgClient, err := wg.CreateClient(conf.IfName)

	var robinRpc robin.RobinRpc
	var robinIpc robin.RobinIpc
	var authProvider middleware.AuthRpcProvider

	ctrlServerParams := ctrlserver.NewCtrlServerParams{
		WgClient:     wgClient,
		Conf:         conf,
		AuthProvider: &authProvider,
		CtrlProvider: &robinRpc,
	}

	ctrlServer, err := ctrlserver.NewCtrlServer(&ctrlServerParams)
	authProvider.Manager = ctrlServer.ConnectionServer.JwtManager
	robinRpc.Server = ctrlServer
	robinIpc.Server = ctrlServer

	if err != nil {
		logging.ErrorLog.Fatalln(err.Error())
	}

	log.Println("Running IPC Handler")

	go ipc.RunIpcHandler(&robinIpc)

	err = ctrlServer.ConnectionServer.Listen()

	if err != nil {
		logging.ErrorLog.Fatalln(err.Error())
	}

	defer wgClient.Close()
}
