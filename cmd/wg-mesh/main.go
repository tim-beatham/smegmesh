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

func createNewMesh(client *ipcRpc.Client) string {
	var reply string
	err := client.Call("Mesh.CreateNewMesh", "", &reply)

	if err != nil {
		return err.Error()
	}

	return reply
}

func listMeshes(client *ipcRpc.Client) {
	var reply map[string]ctrlserver.Mesh

	err := client.Call("Mesh.ListMeshes", "", &reply)

	if err != nil {
		err.Error()
		return
	}

	for sharedKey := range reply {
		fmt.Println(sharedKey)
	}
}

func joinMesh(client *ipcRpc.Client, meshId string, ipAddress string) string {
	var reply string

	args := ipc.JoinMeshArgs{MeshId: meshId, IpAdress: ipAddress}

	err := client.Call("Mesh.JoinMesh", &args, &reply)

	if err != nil {
		return err.Error()
	}

	return reply
}

func getMesh(client *ipcRpc.Client, meshId string) {
	reply := new(ipc.GetMeshReply)

	err := client.Call("Mesh.GetMesh", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, node := range reply.Nodes {
		fmt.Println("Public Key: " + node.PublicKey)
		fmt.Println("WireGuard Endpoint: " + node.HostEndpoint)
		fmt.Println("Control Endpoint: " + node.WgEndpoint)
		fmt.Println("Wg IP: " + node.WgHost)
		fmt.Println("---")
	}
}

func main() {
	parser := argparse.NewParser("wg-mesh",
		"wg-mesh Manipulate WireGuard meshes")

	newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
	listMeshCmd := parser.NewCommand("list-meshes", "List meshes the node is connected to")
	joinMeshCmd := parser.NewCommand("join-mesh", "Join a mesh network")
	getMeshCmd := parser.NewCommand("get-mesh", "Get a mesh network")

	var meshId *string = joinMeshCmd.String("m", "mesh", &argparse.Options{Required: true})
	var ipAddress *string = joinMeshCmd.String("i", "ip", &argparse.Options{Required: true})

	var getMeshId *string = getMeshCmd.String("m", "mesh", &argparse.Options{Required: true})

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
		fmt.Println(createNewMesh(client))
	}

	if listMeshCmd.Happened() {
		listMeshes(client)
	}

	if joinMeshCmd.Happened() {
		fmt.Println(joinMesh(client, *meshId, *ipAddress))
	}

	if getMeshCmd.Happened() {
		getMesh(client, *getMeshId)
	}
}
