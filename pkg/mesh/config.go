package mesh

import (
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/route"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// MeshConfigApplyer abstracts applying the mesh configuration
type MeshConfigApplyer interface {
	ApplyConfig() error
	RemovePeers(meshId string) error
	SetMeshManager(manager MeshManager)
}

// WgMeshConfigApplyer applies WireGuard configuration
type WgMeshConfigApplyer struct {
	meshManager    MeshManager
	config         *conf.WgMeshConfiguration
	routeInstaller route.RouteInstaller
}

type routeNode struct {
	gateway string
	route   Route
}

func (m *WgMeshConfigApplyer) convertMeshNode(node MeshNode, device *wgtypes.Device,
	peerToClients map[string][]net.IPNet,
	routes map[string][]routeNode) (*wgtypes.PeerConfig, error) {

	endpoint, err := net.ResolveUDPAddr("udp", node.GetWgEndpoint())

	if err != nil {
		return nil, err
	}

	pubKey, err := node.GetPublicKey()

	if err != nil {
		return nil, err
	}

	allowedips := make([]net.IPNet, 1)
	allowedips[0] = *node.GetWgHost()

	clients, ok := peerToClients[node.GetWgHost().String()]

	if ok {
		allowedips = append(allowedips, clients...)
	}

	for _, route := range node.GetRoutes() {
		bestRoutes := routes[route.GetDestination().String()]
		var pickedRoute routeNode

		if len(bestRoutes) == 1 {
			pickedRoute = bestRoutes[0]
		} else if len(bestRoutes) > 1 {
			keyFunc := func(mn MeshNode) int {
				pubKey, _ := mn.GetPublicKey()
				return lib.HashString(pubKey.String())
			}

			bucketFunc := func(rn routeNode) int {
				return lib.HashString(rn.gateway)
			}

			// Else there is more than one candidate so consistently hash
			pickedRoute = lib.ConsistentHash(bestRoutes, node, bucketFunc, keyFunc)
		}

		if pickedRoute.gateway == pubKey.String() {
			allowedips = append(allowedips, *pickedRoute.route.GetDestination())
		}
	}

	keepAlive := time.Duration(m.config.KeepAliveWg) * time.Second

	existing := slices.IndexFunc(device.Peers, func(p wgtypes.Peer) bool {
		pubKey, _ := node.GetPublicKey()
		return p.PublicKey.String() == pubKey.String()
	})

	if existing != -1 {
		endpoint = device.Peers[existing].Endpoint
	}

	peerConfig := wgtypes.PeerConfig{
		PublicKey:                   pubKey,
		Endpoint:                    endpoint,
		AllowedIPs:                  allowedips,
		PersistentKeepaliveInterval: &keepAlive,
		ReplaceAllowedIPs:           true,
	}

	return &peerConfig, nil
}

// getRoutes: finds the routes with the least hop distance. If more than one route exists
// consistently hash to evenly spread the distribution of traffic
func (m *WgMeshConfigApplyer) getRoutes(meshProvider MeshProvider) map[string][]routeNode {
	mesh, _ := meshProvider.GetMesh()

	routes := make(map[string][]routeNode)

	meshPrefixes := lib.Map(lib.MapValues(m.meshManager.GetMeshes()), func(mesh MeshProvider) *net.IPNet {
		ula := &ip.ULABuilder{}
		ipNet, _ := ula.GetIPNet(mesh.GetMeshId())

		return ipNet
	})

	for _, node := range mesh.GetNodes() {
		pubKey, _ := node.GetPublicKey()

		for _, route := range node.GetRoutes() {
			if lib.Contains(meshPrefixes, func(prefix *net.IPNet) bool {
				return prefix.Contains(route.GetDestination().IP)
			}) {
				continue
			}

			destination := route.GetDestination().String()
			otherRoute, ok := routes[destination]

			rn := routeNode{
				gateway: pubKey.String(),
				route:   route,
			}

			if !ok {
				otherRoute = make([]routeNode, 1)
				otherRoute[0] = rn
				routes[destination] = otherRoute
			} else if route.GetHopCount() < otherRoute[0].route.GetHopCount() {
				otherRoute[0] = rn
			} else if otherRoute[0].route.GetHopCount() == route.GetHopCount() {
				logging.Log.WriteInfof("Other Route Hop: %d", otherRoute[0].route.GetHopCount())
				logging.Log.WriteInfof("Route gateway %s, route hop %d", rn.gateway, route.GetHopCount())
				routes[destination] = append(otherRoute, rn)
			}
		}
	}

	return routes
}

func (m *WgMeshConfigApplyer) updateWgConf(mesh MeshProvider) error {
	snap, err := mesh.GetMesh()

	if err != nil {
		return err
	}

	nodes := lib.MapValues(snap.GetNodes())
	peerConfigs := make([]wgtypes.PeerConfig, len(nodes))

	peers := lib.Filter(nodes, func(mn MeshNode) bool {
		return mn.GetType() == conf.PEER_ROLE
	})

	var count int = 0

	self, err := m.meshManager.GetSelf(mesh.GetMeshId())

	if err != nil {
		return err
	}

	peerToClients := make(map[string][]net.IPNet)
	routes := m.getRoutes(mesh)
	installedRoutes := make([]lib.Route, 0)

	for _, n := range nodes {
		if NodeEquals(n, self) {
			continue
		}

		if n.GetType() == conf.CLIENT_ROLE && len(peers) > 0 && self.GetType() == conf.CLIENT_ROLE {
			hashFunc := func(mn MeshNode) int {
				return lib.HashString(mn.GetWgHost().String())
			}
			peer := lib.ConsistentHash(peers, n, hashFunc, hashFunc)

			clients, ok := peerToClients[peer.GetWgHost().String()]

			if !ok {
				clients = make([]net.IPNet, 0)
				peerToClients[peer.GetWgHost().String()] = clients
			}

			peerToClients[peer.GetWgHost().String()] = append(clients, *n.GetWgHost())
			continue
		}

		dev, _ := mesh.GetDevice()
		peer, err := m.convertMeshNode(n, dev, peerToClients, routes)

		if err != nil {
			return err
		}

		for _, route := range peer.AllowedIPs {
			ula := &ip.ULABuilder{}
			ipNet, _ := ula.GetIPNet(mesh.GetMeshId())

			if !ipNet.Contains(route.IP) {
				installedRoutes = append(installedRoutes, lib.Route{
					Gateway:     n.GetWgHost().IP,
					Destination: route,
				})
			}
		}

		peerConfigs[count] = *peer
		count++
	}

	cfg := wgtypes.Config{
		Peers:        peerConfigs,
		ReplacePeers: true,
	}

	dev, err := mesh.GetDevice()

	if err != nil {
		return err
	}

	err = m.meshManager.GetClient().ConfigureDevice(dev.Name, cfg)

	if err != nil {
		return err
	}

	return m.routeInstaller.InstallRoutes(dev.Name, installedRoutes...)
}

func (m *WgMeshConfigApplyer) ApplyConfig() error {
	for _, mesh := range m.meshManager.GetMeshes() {
		err := m.updateWgConf(mesh)

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *WgMeshConfigApplyer) RemovePeers(meshId string) error {
	mesh := m.meshManager.GetMesh(meshId)

	if mesh == nil {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	dev, err := mesh.GetDevice()

	if err != nil {
		return err
	}

	m.meshManager.GetClient().ConfigureDevice(dev.Name, wgtypes.Config{
		Peers: make([]wgtypes.PeerConfig, 0),
	})

	return nil
}

func (m *WgMeshConfigApplyer) SetMeshManager(manager MeshManager) {
	m.meshManager = manager
}

func NewWgMeshConfigApplyer(config *conf.WgMeshConfiguration) MeshConfigApplyer {
	return &WgMeshConfigApplyer{
		config:         config,
		routeInstaller: route.NewRouteInstaller(),
	}
}
