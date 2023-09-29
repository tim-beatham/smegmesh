package main

import (
	"log"
	"net"

	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/robin"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	wg "github.com/tim-beatham/wgmesh/pkg/wg"
)

const ifName = "wgmesh"

func main() {
	wgClient, err := wg.CreateClient(ifName)

	if err != nil {
		log.Fatalf("Could not create interface %s\n", ifName)
	}

	ctrlServer := ctrlserver.NewCtrlServer(wgClient, "wgmesh")

	log.Println("Running IPC Handler")

	robinIpc := robin.NewRobinIpc(ctrlServer)
	robinRpc := robin.NewRobinRpc(ctrlServer)

	go ipc.RunIpcHandler(robinIpc)
	grpc := rpc.NewRpcServer(*&robinRpc)

	lis, err := net.Listen("tcp", ":8080")
	if err := grpc.Serve(lis); err != nil {
		log.Fatal(err.Error())
	}

	defer wgClient.Close()
}
