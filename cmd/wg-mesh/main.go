package main

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/akamensky/argparse"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

func createNewMesh(client *rpc.Client) {
	var reply string
	err := client.Call("Mesh.CreateNewMesh", "", &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func listMeshes(client *rpc.Client) {
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

func joinMesh(client *rpc.Client, meshId string, ipAddress string) {
	fmt.Println(meshId + " " + ipAddress)
}

func main() {
	parser := argparse.NewParser("wg-mesh",
		"wg-mesh Manipulate WireGuard meshes")

	newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
	listMeshCmd := parser.NewCommand("list-meshes", "List meshes the node is connected to")
	joinMeshCmd := parser.NewCommand("join-mesh", "Join a mesh network")

	var meshId *string = joinMeshCmd.StringPositional(&argparse.Options{Required: true})
	var ipAddress *string = joinMeshCmd.StringPositional(&argparse.Options{Required: true})

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	client, err := rpc.DialHTTP("unix", SockAddr)
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
