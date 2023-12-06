package mesh

import (
	"fmt"
	"net"
	"slices"
	"strings"
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

	clients, ok := peerToClients[pubKey.String()]

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

// getCorrespondignPeer: gets the peer corresponding to the client
func (m *WgMeshConfigApplyer) getCorrespondingPeer(peers []MeshNode, client MeshNode) MeshNode {
	hashFunc := func(mn MeshNode) int {
		pubKey, _ := mn.GetPublicKey()
		return lib.HashString(pubKey.String())
	}

	peer := lib.ConsistentHash(peers, client, hashFunc, hashFunc)
	return peer
}

func (m *WgMeshConfigApplyer) getClientConfig(mesh MeshProvider, peers []MeshNode, clients []MeshNode) (*wgtypes.Config, error) {
	self, err := m.meshManager.GetSelf(mesh.GetMeshId())

	if err != nil {
		return nil, err
	}

	peer := m.getCorrespondingPeer(peers, self)

	pubKey, _ := peer.GetPublicKey()

	keepAlive := time.Duration(m.config.KeepAliveWg) * time.Second
	endpoint, err := net.ResolveUDPAddr("udp", peer.GetWgEndpoint())

	if err != nil {
		return nil, err
	}

	allowedips := make([]net.IPNet, 1)
	_, ipnet, _ := net.ParseCIDR("::/0")
	allowedips[0] = *ipnet

	peerCfgs := make([]wgtypes.PeerConfig, 1)

	peerCfgs[0] = wgtypes.PeerConfig{
		PublicKey:                   pubKey,
		Endpoint:                    endpoint,
		PersistentKeepaliveInterval: &keepAlive,
		AllowedIPs:                  allowedips,
	}

	cfg := wgtypes.Config{
		Peers: peerCfgs,
	}

	return &cfg, err
}

func (m *WgMeshConfigApplyer) getPeerConfig(mesh MeshProvider, peers []MeshNode, clients []MeshNode, dev *wgtypes.Device) (*wgtypes.Config, error) {
	peerToClients := make(map[string][]net.IPNet)
	routes := m.getRoutes(mesh)
	installedRoutes := make([]lib.Route, 0)
	peerConfigs := make([]wgtypes.PeerConfig, 0)
	self, err := m.meshManager.GetSelf(mesh.GetMeshId())

	if err != nil {
		return nil, err
	}

	for _, n := range clients {
		if len(peers) > 0 {
			peer := m.getCorrespondingPeer(peers, n)
			pubKey, _ := peer.GetPublicKey()
			clients, ok := peerToClients[pubKey.String()]

			if !ok {
				clients = make([]net.IPNet, 0)
				peerToClients[pubKey.String()] = clients
			}

			peerToClients[pubKey.String()] = append(clients, *n.GetWgHost())

			if NodeEquals(self, peer) {
				cfg, err := m.convertMeshNode(n, dev, peerToClients, routes)

				if err != nil {
					return nil, err
				}

				peerConfigs = append(peerConfigs, *cfg)
			}
		}
	}

	for _, n := range peers {
		if NodeEquals(n, self) {
			continue
		}

		peer, err := m.convertMeshNode(n, dev, peerToClients, routes)

		if err != nil {
			return nil, err
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

		peerConfigs = append(peerConfigs, *peer)
	}

	cfg := wgtypes.Config{
		Peers:        peerConfigs,
		ReplacePeers: true,
	}

	err = m.routeInstaller.InstallRoutes(dev.Name, installedRoutes...)
	return &cfg, err
}

func (m *WgMeshConfigApplyer) updateWgConf(mesh MeshProvider) error {
	snap, err := mesh.GetMesh()

	if err != nil {
		return err
	}

	nodes := lib.MapValues(snap.GetNodes())
	dev, _ := mesh.GetDevice()

	slices.SortFunc(nodes, func(a, b MeshNode) int {
		return strings.Compare(string(a.GetType()), string(b.GetType()))
	})

	peers := lib.Filter(nodes, func(mn MeshNode) bool {
		return mn.GetType() == conf.PEER_ROLE
	})

	clients := lib.Filter(nodes, func(mn MeshNode) bool {
		return mn.GetType() == conf.CLIENT_ROLE
	})

	self, err := m.meshManager.GetSelf(mesh.GetMeshId())

	if err != nil {
		return err
	}

	var cfg *wgtypes.Config = nil

	switch self.GetType() {
	case conf.PEER_ROLE:
		cfg, err = m.getPeerConfig(mesh, peers, clients, dev)
	case conf.CLIENT_ROLE:
		cfg, err = m.getClientConfig(mesh, peers, clients)
	}

	if err != nil {
		return err
	}

	err = m.meshManager.GetClient().ConfigureDevice(dev.Name, *cfg)

	if err != nil {
		return err
	}

	return nil
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
		Peers:        make([]wgtypes.PeerConfig, 0),
		ReplacePeers: true,
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
