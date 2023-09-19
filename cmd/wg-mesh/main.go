package main

import (
	"fmt"
	ipcRpc "net/rpc"
	"os"

	"github.com/akamensky/argparse"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

func createNewMesh(client *ipcRpc.Client) {
	var reply string
	err := client.Call("Mesh.CreateNewMesh", "", &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func listMeshes(client *ipcRpc.Client) {
	var reply map[string]ctrlserver.Mesh

	err := client.Call("Mesh.ListMeshes", "", &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for sharedKey := range reply {
		fmt.Println(sharedKey)
	}
}

func joinMesh(client *ipcRpc.Client, meshId string, ipAddress string) {
	var reply string

	args := ipc.JoinMeshArgs{MeshId: meshId, IpAdress: ipAddress}

	err := client.Call("Mesh.JoinMesh", &args, &reply)

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

	var meshId *string = joinMeshCmd.String("m", "mesh", &argparse.Options{Required: true})
	var ipAddress *string = joinMeshCmd.String("i", "ip", &argparse.Options{Required: true})

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
		createNewMesh(client)
	}

	if listMeshCmd.Happened() {
		listMeshes(client)
	}

	if joinMeshCmd.Happened() {
		joinMesh(client, *meshId, *ipAddress)
	}
}
