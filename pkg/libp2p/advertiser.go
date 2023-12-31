package main

import (
	"bufio"
	"context"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

const PROTOCOL_ID = "/smegmesh/1.0"

type Advertiser interface {
}

type Libp2pAdvertiser struct {
	host host.Host
	dht  *dht.IpfsDHT
	*drouting.RoutingDiscovery
}

func readData(bf *bufio.ReadWriter) {
}

func writeData(bf *bufio.ReadWriter) {
	bf.Write([]byte("My name is Tim"))
}

func handleStream(stream network.Stream) {
	logging.Log.WriteInfof("Received a new stream!")

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)
}

func NewLibP2PAdvertiser() (Advertiser, error) {
	logging.Log.WriteInfof("setting up")

	addrs := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0")
	host, err := libp2p.New(addrs)

	if err != nil {
		return nil, err
	}

	logging.Log.WriteInfof("Host created. We are: ", host.ID())

	ctx := context.Background()

	logging.Log.WriteInfof("creating DHT")

	kDHT, err := dht.New(ctx, host)

	if err != nil {
		return nil, err
	}

	logging.Log.WriteInfof("bootstrapping the DHT")
	if err := kDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				logging.Log.WriteErrorf(err.Error())
			} else {
				logging.Log.WriteInfof("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	logging.Log.WriteInfof("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kDHT)
	dutil.Advertise(ctx, routingDiscovery, "bobmarley")
	logging.Log.WriteInfof("Successfully announced!")

	select {}

	return nil, err
}

func main() {
	_, err := NewLibP2PAdvertiser()

	if err != nil {
		logging.Log.WriteInfof(err.Error())
	}
}
