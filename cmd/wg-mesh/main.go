package main

import (
    "fmt"
    "os"
    "github.com/akamensky/argparse"
)

func main () {
    parser := argparse.NewParser("wg-mesh",
            "wg-mesh Manipulate WireGuard meshes")

    newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
    err := parser.Parse(os.Args)

    if err != nil {
        fmt.Print(parser.Usage(err))
        return
    }

    fmt.Println(newMeshCmd.Happened())
}