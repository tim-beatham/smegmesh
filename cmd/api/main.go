package main

import (
	"log"

	"github.com/tim-beatham/wgmesh/pkg/api"
)

func main() {
	apiServer, err := api.NewSmegServer()

	if err != nil {
		log.Fatal(err.Error())
	}

	apiServer.Run(":40000")
}
