package main

import (
	"fmt"
	"net"

	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/ipc"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/rpc"
	wg "github.com/tim-beatham/wgmesh/pkg/wg"
)

func main() {
	wgClient, err := wg.CreateClient("wgmesh")

	if err != nil {
		fmt.Println(err)
		return
	}

	ctrlServer := ctrlserver.NewCtrlServer(wgClient, "wgmesh")

	fmt.Println("Running IPC Handler")
	go ipc.RunIpcHandler(ctrlServer)

	grpc := rpc.NewRpcServer(ctrlServer)

	lis, err := net.Listen("tcp", ":8080")
	if err := grpc.Serve(lis); err != nil {
		fmt.Print(err.Error())
	}
}
