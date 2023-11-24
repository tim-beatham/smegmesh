package main

import (
	"log"

	smegdns "github.com/tim-beatham/wgmesh/pkg/dns"
)

func main() {
	server, err := smegdns.NewDns(53)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer server.Close()
	server.Listen()
}
