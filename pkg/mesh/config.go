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
	hashFunc       func(MeshNode) int
}

type routeNode struct {
	gateway string
	route   Route
}

func (m *WgMeshConfigApplyer) convertMeshNode(node MeshNode, self MeshNode,
	device *wgtypes.Device,
	peerToClients map[string][]net.IPNet,
	routes map[string][]routeNode) (*wgtypes.PeerConfig, error) {

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
			bucketFunc := func(rn routeNode) int {
				return lib.HashString(rn.gateway)
			}

			// Else there is more than one candidate so consistently hash
			pickedRoute = lib.ConsistentHash(bestRoutes, self, bucketFunc, m.hashFunc)
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

	endpoint, err := net.ResolveUDPAddr("udp", node.GetWgEndpoint())

	if err != nil {
		return nil, err
	}

	// Don't override the existing IP in case it already exists
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

	peers := lib.Filter(lib.MapValues(mesh.GetNodes()), func(p MeshNode) bool {
		return p.GetType() == conf.PEER_ROLE
	})

	meshPrefixes := lib.Map(lib.MapValues(m.meshManager.GetMeshes()), func(mesh MeshProvider) *net.IPNet {
		ula := &ip.ULABuilder{}
		ipNet, _ := ula.GetIPNet(mesh.GetMeshId())
		return ipNet
	})

	for _, node := range mesh.GetNodes() {
		pubKey, _ := node.GetPublicKey()

		for _, route := range node.GetRoutes() {
			if lib.Contains(meshPrefixes, func(prefix *net.IPNet) bool {
				v6Default, _, _ := net.ParseCIDR("::/0")
				v4Default, _, _ := net.ParseCIDR("0.0.0.0/0")

				if (prefix.IP.Equal(v6Default) || prefix.IP.Equal(v4Default)) && m.config.AdvertiseDefaultRoute {
					return true
				}

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

			// Client's only acessible by another peer
			if node.GetType() == conf.CLIENT_ROLE {
				peer := m.getCorrespondingPeer(peers, node)
				self, _ := m.meshManager.GetSelf(meshProvider.GetMeshId())

				// If the node isn't the self use that peer as the gateway
				if !NodeEquals(peer, self) {
					peerPub, _ := peer.GetPublicKey()
					rn.gateway = peerPub.String()
					rn.route = &RouteStub{
						Destination: rn.route.GetDestination(),
						HopCount:    rn.route.GetHopCount() + 1,
						// Append the path to this peer
						Path: append(rn.route.GetPath(), peer.GetWgHost().IP.String()),
					}
				}
			}

			if !ok {
				otherRoute = make([]routeNode, 1)
				otherRoute[0] = rn
				routes[destination] = otherRoute
			} else if route.GetHopCount() < otherRoute[0].route.GetHopCount() {
				otherRoute[0] = rn
			} else if otherRoute[0].route.GetHopCount() == route.GetHopCount() {
				routes[destination] = append(otherRoute, rn)
			}
		}
	}

	return routes
}

// getCorrespondignPeer: gets the peer corresponding to the client
func (m *WgMeshConfigApplyer) getCorrespondingPeer(peers []MeshNode, client MeshNode) MeshNode {
	peer := lib.ConsistentHash(peers, client, m.hashFunc, m.hashFunc)
	return peer
}

func (m *WgMeshConfigApplyer) getPeerCfgsToRemove(dev *wgtypes.Device, newPeers []wgtypes.PeerConfig) []wgtypes.PeerConfig {
	peers := dev.Peers
	peers = lib.Filter(peers, func(p1 wgtypes.Peer) bool {
		return !lib.Contains(newPeers, func(p2 wgtypes.PeerConfig) bool {
			return p1.PublicKey.String() == p2.PublicKey.String()
		})
	})

	return lib.Map(peers, func(p wgtypes.Peer) wgtypes.PeerConfig {
		return wgtypes.PeerConfig{
			PublicKey: p.PublicKey,
			Remove:    true,
		}
	})
}

type GetConfigParams struct {
	mesh    MeshProvider
	peers   []MeshNode
	clients []MeshNode
	dev     *wgtypes.Device
	routes  map[string][]routeNode
}

func (m *WgMeshConfigApplyer) getClientConfig(params *GetConfigParams) (*wgtypes.Config, error) {
	self, err := m.meshManager.GetSelf(params.mesh.GetMeshId())
	ula := &ip.ULABuilder{}
	meshNet, _ := ula.GetIPNet(params.mesh.GetMeshId())

	routesForMesh := lib.Map(lib.MapValues(params.routes), func(rns []routeNode) []routeNode {
		return lib.Filter(rns, func(rn routeNode) bool {
			ip, _, _ := net.ParseCIDR(rn.gateway)
			return meshNet.Contains(ip)
		})
	})

	routes := lib.Map(routesForMesh, func(rs []routeNode) net.IPNet {
		return *rs[0].route.GetDestination()
	})
	routes = append(routes, *meshNet)

	if err != nil {
		return nil, err
	}

	peer := m.getCorrespondingPeer(params.peers, self)
	pubKey, _ := peer.GetPublicKey()
	keepAlive := time.Duration(m.config.KeepAliveWg) * time.Second
	endpoint, err := net.ResolveUDPAddr("udp", peer.GetWgEndpoint())

	if err != nil {
		return nil, err
	}

	peerCfgs := make([]wgtypes.PeerConfig, 1)

	peerCfgs[0] = wgtypes.PeerConfig{
		PublicKey:                   pubKey,
		Endpoint:                    endpoint,
		PersistentKeepaliveInterval: &keepAlive,
		AllowedIPs:                  routes,
		ReplaceAllowedIPs:           true,
	}

	installedRoutes := make([]lib.Route, 0)

	for _, route := range peerCfgs[0].AllowedIPs {
		installedRoutes = append(installedRoutes, lib.Route{
			Gateway:     peer.GetWgHost().IP,
			Destination: route,
		})
	}

	cfg := wgtypes.Config{
		Peers: peerCfgs,
	}

	m.routeInstaller.InstallRoutes(params.dev.Name, installedRoutes...)
	return &cfg, err
}

func (m *WgMeshConfigApplyer) getRoutesToInstall(wgNode *wgtypes.PeerConfig, mesh MeshProvider, node MeshNode) []lib.Route {
	routes := make([]lib.Route, 0)

	for _, route := range wgNode.AllowedIPs {
		ula := &ip.ULABuilder{}
		ipNet, _ := ula.GetIPNet(mesh.GetMeshId())

		_, defaultRoute, _ := net.ParseCIDR("::/0")

		if !ipNet.Contains(route.IP) && !ipNet.IP.Equal(defaultRoute.IP) {
			routes = append(routes, lib.Route{
				Gateway:     node.GetWgHost().IP,
				Destination: route,
			})
		}
	}

	return routes
}

func (m *WgMeshConfigApplyer) getPeerConfig(params *GetConfigParams) (*wgtypes.Config, error) {
	peerToClients := make(map[string][]net.IPNet)
	installedRoutes := make([]lib.Route, 0)
	peerConfigs := make([]wgtypes.PeerConfig, 0)
	self, err := m.meshManager.GetSelf(params.mesh.GetMeshId())

	if err != nil {
		return nil, err
	}

	for _, n := range params.clients {
		if len(params.peers) > 0 {
			peer := m.getCorrespondingPeer(params.peers, n)
			pubKey, _ := peer.GetPublicKey()
			clients, ok := peerToClients[pubKey.String()]

			if !ok {
				clients = make([]net.IPNet, 0)
				peerToClients[pubKey.String()] = clients
			}

			peerToClients[pubKey.String()] = append(clients, *n.GetWgHost())

			if NodeEquals(self, peer) {
				cfg, err := m.convertMeshNode(n, self, params.dev, peerToClients, params.routes)

				if err != nil {
					return nil, err
				}

				installedRoutes = append(installedRoutes, m.getRoutesToInstall(cfg, params.mesh, n)...)
				peerConfigs = append(peerConfigs, *cfg)
			}
		}
	}

	for _, n := range params.peers {
		if NodeEquals(n, self) {
			continue
		}

		peer, err := m.convertMeshNode(n, self, params.dev, peerToClients, params.routes)

		if err != nil {
			return nil, err
		}

		installedRoutes = append(installedRoutes, m.getRoutesToInstall(peer, params.mesh, n)...)
		peerConfigs = append(peerConfigs, *peer)
	}

	cfg := wgtypes.Config{
		Peers: peerConfigs,
	}

	err = m.routeInstaller.InstallRoutes(params.dev.Name, installedRoutes...)
	return &cfg, err
}

func (m *WgMeshConfigApplyer) updateWgConf(mesh MeshProvider, routes map[string][]routeNode) error {
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

	configParams := &GetConfigParams{
		mesh:    mesh,
		peers:   peers,
		clients: clients,
		dev:     dev,
		routes:  routes,
	}

	switch self.GetType() {
	case conf.PEER_ROLE:
		cfg, err = m.getPeerConfig(configParams)
	case conf.CLIENT_ROLE:
		cfg, err = m.getClientConfig(configParams)
	}

	if err != nil {
		return err
	}

	toRemove := m.getPeerCfgsToRemove(dev, cfg.Peers)
	cfg.Peers = append(cfg.Peers, toRemove...)

	err = m.meshManager.GetClient().ConfigureDevice(dev.Name, *cfg)

	if err != nil {
		return err
	}

	return nil
}

func (m *WgMeshConfigApplyer) getAllRoutes() map[string][]routeNode {
	allRoutes := make(map[string][]routeNode)

	for _, mesh := range m.meshManager.GetMeshes() {
		routes := m.getRoutes(mesh)

		for destination, route := range routes {
			_, ok := allRoutes[destination]

			if !ok {
				allRoutes[destination] = route
				continue
			}

			if allRoutes[destination][0].route.GetHopCount() == route[0].route.GetHopCount() {
				allRoutes[destination] = append(allRoutes[destination], route...)
			} else if route[0].route.GetHopCount() < allRoutes[destination][0].route.GetHopCount() {
				allRoutes[destination] = route
			}
		}
	}

	return allRoutes
}

func (m *WgMeshConfigApplyer) ApplyConfig() error {
	allRoutes := m.getAllRoutes()

	for _, mesh := range m.meshManager.GetMeshes() {
		err := m.updateWgConf(mesh, allRoutes)

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
		hashFunc: func(mn MeshNode) int {
			pubKey, _ := mn.GetPublicKey()
			return lib.HashString(pubKey.String())
		},
	}
}
