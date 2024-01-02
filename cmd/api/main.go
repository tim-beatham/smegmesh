package main

import (
	"log"

	"github.com/tim-beatham/smegmesh/pkg/api"
)

func main() {
	apiServer, err := api.NewSmegServer(api.ApiServerConf{
		WordsFile: "./cmd/api/words.txt",
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	apiServer.Run(":8080")
}
