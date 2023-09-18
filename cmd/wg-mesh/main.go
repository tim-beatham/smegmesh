package main

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/akamensky/argparse"
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

func main() {
	parser := argparse.NewParser("wg-mesh",
		"wg-mesh Manipulate WireGuard meshes")

	newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
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
}
