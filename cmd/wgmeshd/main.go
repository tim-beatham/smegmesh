package main

import (
	"log"
	"net"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/middleware"
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

	conf, err := conf.ParseConfiguration("./configuration.yaml")

	newConnParams := conn.NewConnectionsParams{
		CertificatePath:      conf.CertificatePath,
		PrivateKey:           conf.PrivateKeyPath,
		SkipCertVerification: conf.SkipCertVerification,
	}

	conn, err := conn.NewConnection(&newConnParams)

	if err != nil {
		return
	}

	ctrlServer := ctrlserver.NewCtrlServer(wgClient, conn, "wgmesh")

	log.Println("Running IPC Handler")

	robinIpc := robin.NewRobinIpc(ctrlServer)
	robinRpc := robin.NewRobinRpc(ctrlServer)

	go ipc.RunIpcHandler(robinIpc)

	grpc := conn.Listen(ctrlServer.JwtManager.GetAuthInterceptor())
	rpc.NewRpcServer(grpc, robinRpc, middleware.NewAuthProvider(ctrlServer))

	lis, err := net.Listen("tcp", ":8080")
	if err := grpc.Serve(lis); err != nil {
		log.Fatal(err.Error())
	}

	defer wgClient.Close()
}
