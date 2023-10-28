package main

import (
	"fmt"
	ipcRpc "net/rpc"
	"os"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

type CreateMeshParams struct {
	Client   *ipcRpc.Client
	IfName   string
	WgPort   int
	Endpoint string
}

func createMesh(args *CreateMeshParams) string {
	var reply string
	newMeshParams := ipc.NewMeshArgs{
		IfName:   args.IfName,
		WgPort:   args.WgPort,
		Endpoint: args.Endpoint,
	}

	err := args.Client.Call("IpcHandler.CreateMesh", &newMeshParams, &reply)

	if err != nil {
		return err.Error()
	}

	return reply
}

func listMeshes(client *ipcRpc.Client) {
	reply := new(ipc.ListMeshReply)

	err := client.Call("IpcHandler.ListMeshes", "", &reply)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return
	}

	for _, meshId := range reply.Meshes {
		fmt.Println(meshId)
	}
}

type JoinMeshParams struct {
	Client    *ipcRpc.Client
	MeshId    string
	IpAddress string
	IfName    string
	WgPort    int
	Endpoint  string
}

func joinMesh(params *JoinMeshParams) string {
	var reply string

	args := ipc.JoinMeshArgs{
		MeshId:   params.MeshId,
		IpAdress: params.IpAddress,
		IfName:   params.IfName,
		Port:     params.WgPort,
	}

	err := params.Client.Call("IpcHandler.JoinMesh", &args, &reply)

	if err != nil {
		return err.Error()
	}

	return reply
}

func getMesh(client *ipcRpc.Client, meshId string) {
	reply := new(ipc.GetMeshReply)

	err := client.Call("IpcHandler.GetMesh", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, node := range reply.Nodes {
		fmt.Println("Public Key: " + node.PublicKey)
		fmt.Println("Control Endpoint: " + node.HostEndpoint)
		fmt.Println("WireGuard Endpoint: " + node.WgEndpoint)
		fmt.Println("Wg IP: " + node.WgHost)
		fmt.Println(fmt.Sprintf("Timestamp: %s", time.Unix(node.Timestamp, 0).String()))

		advertiseRoutes := strings.Join(node.Routes, ",")
		fmt.Printf("Routes: %s\n", advertiseRoutes)

		fmt.Println("---")
	}
}

func leaveMesh(client *ipcRpc.Client, meshId string) {
	var reply string

	err := client.Call("IpcHandler.LeaveMesh", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func enableInterface(client *ipcRpc.Client, meshId string) {
	var reply string

	err := client.Call("IpcHandler.EnableInterface", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func getGraph(client *ipcRpc.Client, meshId string) {
	var reply string

	err := client.Call("IpcHandler.GetDOT", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func main() {
	parser := argparse.NewParser("wg-mesh",
		"wg-mesh Manipulate WireGuard meshes")

	newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
	listMeshCmd := parser.NewCommand("list-meshes", "List meshes the node is connected to")
	joinMeshCmd := parser.NewCommand("join-mesh", "Join a mesh network")
	getMeshCmd := parser.NewCommand("get-mesh", "Get a mesh network")
	enableInterfaceCmd := parser.NewCommand("enable-interface", "Enable A Specific Mesh Interface")
	getGraphCmd := parser.NewCommand("get-graph", "Convert a mesh into DOT format")
	leaveMeshCmd := parser.NewCommand("leave-mesh", "Leave a mesh network")

	var newMeshIfName *string = newMeshCmd.String("f", "ifname", &argparse.Options{Required: true})
	var newMeshPort *int = newMeshCmd.Int("p", "wgport", &argparse.Options{Required: true})
	var newMeshEndpoint *string = newMeshCmd.String("e", "endpoint", &argparse.Options{})

	var joinMeshId *string = joinMeshCmd.String("m", "mesh", &argparse.Options{Required: true})
	var joinMeshIpAddress *string = joinMeshCmd.String("i", "ip", &argparse.Options{Required: true})
	var joinMeshIfName *string = joinMeshCmd.String("f", "ifname", &argparse.Options{Required: true})
	var joinMeshPort *int = joinMeshCmd.Int("p", "wgport", &argparse.Options{Required: true})
	var joinMeshEndpoint *string = joinMeshCmd.String("e", "endpoint", &argparse.Options{})

	var getMeshId *string = getMeshCmd.String("m", "mesh", &argparse.Options{Required: true})
	var enableInterfaceMeshId *string = enableInterfaceCmd.String("m", "mesh", &argparse.Options{Required: true})
	var getGraphMeshId *string = getGraphCmd.String("m", "mesh", &argparse.Options{Required: true})

	var leaveMeshMeshId *string = leaveMeshCmd.String("m", "mesh", &argparse.Options{Required: true})

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	client, err := ipcRpc.DialHTTP("unix", SockAddr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if newMeshCmd.Happened() {
		fmt.Println(createMesh(&CreateMeshParams{
			Client:   client,
			IfName:   *newMeshIfName,
			WgPort:   *newMeshPort,
			Endpoint: *newMeshEndpoint,
		}))
	}

	if listMeshCmd.Happened() {
		listMeshes(client)
	}

	if joinMeshCmd.Happened() {
		fmt.Println(joinMesh(&JoinMeshParams{
			Client:    client,
			IfName:    *joinMeshIfName,
			WgPort:    *joinMeshPort,
			IpAddress: *joinMeshIpAddress,
			MeshId:    *joinMeshId,
			Endpoint:  *joinMeshEndpoint,
		}))
	}

	if getMeshCmd.Happened() {
		getMesh(client, *getMeshId)
	}

	if getGraphCmd.Happened() {
		getGraph(client, *getGraphMeshId)
	}

	if enableInterfaceCmd.Happened() {
		enableInterface(client, *enableInterfaceMeshId)
	}

	if leaveMeshCmd.Happened() {
		leaveMesh(client, *leaveMeshMeshId)
	}
}
