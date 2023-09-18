package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	meshtypes "github.com/tim-beatham/wgmesh/pkg/wg-mesh"
)

func main() {
	parser := argparse.NewParser("wg-mesh",
		"wg-mesh Manipulate WireGuard meshes")

	newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	if newMeshCmd.Happened() {
		mesh, err := meshtypes.NewWgMesh()

		if err != nil {
			fmt.Println("Could not generate new WgMesh")
		} else {
			fmt.Println(mesh.SharedKey.String())
		}
	}
}
