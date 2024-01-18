package mesh

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/ip"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	"github.com/tim-beatham/smegmesh/pkg/route"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// MeshConfigApplyer abstracts applying the mesh configuration
type MeshConfigApplyer interface {
	// ApplyConfig: apply the configurtation
	ApplyConfig() error
	// SetMeshManager: sets the associated manager
	SetMeshManager(manager MeshManager)
}

// WgMeshConfigApplyer: applies WireGuard configuration
type WgMeshConfigApplyer struct {
	meshManager    MeshManager
	routeInstaller route.RouteInstaller
	hashFunc       func(MeshNode) int
}

type routeNode struct {
	gateway string
	route   Route
}

type convertMeshNodeParams struct {
	node          MeshNode
	mesh          MeshProvider
	device        *wgtypes.Device
	peerToClients map[string][]net.IPNet
	routes        map[string][]routeNode
}

func (m *WgMeshConfigApplyer) convertMeshNode(params convertMeshNodeParams) (*wgtypes.PeerConfig, error) {
	pubKey, err := params.node.GetPublicKey()

	if err != nil {
		return nil, err
	}

	allowedips := make([]net.IPNet, 1)
	allowedips[0] = *params.node.GetWgHost()

	clients, ok := params.peerToClients[pubKey.String()]

	if ok {
		allowedips = append(allowedips, clients...)
	}

	for _, bestRoutes := range lib.MapValues(params.routes) {
		var pickedRoute routeNode

		if len(bestRoutes) == 1 {
			pickedRoute = bestRoutes[0]
		} else if len(bestRoutes) > 1 {
			bucketFunc := func(rn routeNode) int {
				return lib.HashString(rn.gateway)
			}

			pickedRoute = lib.ConsistentHash(bestRoutes, params.node, bucketFunc, m.hashFunc)
		}

		if pickedRoute.gateway == pubKey.String() {
			allowedips = append(allowedips, *pickedRoute.route.GetDestination())
		}
	}

	config := params.mesh.GetConfiguration()

	var keepAlive time.Duration = time.Duration(0)

	if config.KeepAliveWg != nil {
		keepAlive = time.Duration(*config.KeepAliveWg) * time.Second
	}

	existing := slices.IndexFunc(params.device.Peers, func(p wgtypes.Peer) bool {
		pubKey, _ := params.node.GetPublicKey()
		return p.PublicKey.String() == pubKey.String()
	})

	var endpoint *net.UDPAddr = nil

	if params.node.GetType() == conf.PEER_ROLE {
		endpoint, err = net.ResolveUDPAddr("udp", params.node.GetWgEndpoint())
	}

	if err != nil {
		return nil, err
	}

	// Don't override the existing IP in case it already exists
	if existing != -1 {
		endpoint = params.device.Peers[existing].Endpoint
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
func (m *WgMeshConfigApplyer) getRoutes(meshProvider MeshProvider) (map[string][]routeNode, error) {
	mesh, err := meshProvider.GetMesh()

	if err != nil {
		return nil, err
	}

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
				if prefix.IP.Equal(net.IPv6zero) && *meshProvider.GetConfiguration().AdvertiseDefaultRoute {
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
				self, err := meshProvider.GetNode(m.meshManager.GetPublicKey().String())

				if err != nil {
					return nil, err
				}

				if !NodeEquals(peer, self) {
					peerPub, _ := peer.GetPublicKey()
					rn.gateway = peerPub.String()
					rn.route = &RouteStub{
						Destination: rn.route.GetDestination(),
						Path:        append(rn.route.GetPath(), peer.GetWgHost().IP.String()),
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

	return routes, nil
}

// getCorrespondignPeer: gets the peer corresponding to the client
func (m *WgMeshConfigApplyer) getCorrespondingPeer(peers []MeshNode, client MeshNode) MeshNode {
	peer := lib.ConsistentHash(peers, client, m.hashFunc, m.hashFunc)
	return peer
}

// getPeerCfgsToRemove: remove peer configurations that are no longer in the mesh
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

// getClientConfig: if the node is a client get their configuration
func (m *WgMeshConfigApplyer) getClientConfig(params *GetConfigParams) (*wgtypes.Config, error) {
	ula := &ip.ULABuilder{}
	meshNet, _ := ula.GetIPNet(params.mesh.GetMeshId())

	routesForMesh := lib.Map(lib.MapValues(params.routes), func(rns []routeNode) []routeNode {
		return lib.Filter(rns, func(rn routeNode) bool {
			node, err := params.mesh.GetNode(rn.gateway)
			return node != nil && err == nil
		})
	})

	routesForMesh = lib.Filter(routesForMesh, func(rns []routeNode) bool {
		return len(rns) != 0
	})

	routes := lib.Map(routesForMesh, func(rs []routeNode) net.IPNet {
		return *rs[0].route.GetDestination()
	})
	routes = append(routes, *meshNet)

	self, err := params.mesh.GetNode(m.meshManager.GetPublicKey().String())

	if err != nil {
		return nil, err
	}

	if len(params.peers) == 0 {
		return nil, fmt.Errorf("no peers in the mesh")
	}

	peer := m.getCorrespondingPeer(params.peers, self)
	pubKey, _ := peer.GetPublicKey()

	config := params.mesh.GetConfiguration()

	keepAlive := time.Duration(*config.KeepAliveWg) * time.Second
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
		// Don't install routes that we are directly apart
		// Dont install default route wgctrl handles this for us
		if !meshNet.Contains(route.IP) {
			installedRoutes = append(installedRoutes, lib.Route{
				Gateway:     peer.GetWgHost().IP,
				Destination: route,
			})
		}
	}

	cfg := wgtypes.Config{
		Peers: peerCfgs,
	}

	if params.dev != nil {
		m.routeInstaller.InstallRoutes(params.dev.Name, installedRoutes...)
	}

	return &cfg, err
}

// getRoutesToInstall: work out if the given node is advertising routes that should be installed into the
// RIB
func (m *WgMeshConfigApplyer) getRoutesToInstall(wgNode *wgtypes.PeerConfig, mesh MeshProvider, node MeshNode) []lib.Route {
	routes := make([]lib.Route, 0)

	for _, route := range wgNode.AllowedIPs {
		ula := &ip.ULABuilder{}
		ipNet, _ := ula.GetIPNet(mesh.GetMeshId())

		// Check there is no overlap in network and its not the default route
		if !ipNet.Contains(route.IP) {
			routes = append(routes, lib.Route{
				Gateway:     node.GetWgHost().IP,
				Destination: route,
			})
		}
	}

	return routes
}

// getPeerConfig: creates the WireGuard configuration for a peer
func (m *WgMeshConfigApplyer) getPeerConfig(params *GetConfigParams) (*wgtypes.Config, error) {
	peerToClients := make(map[string][]net.IPNet)
	installedRoutes := make([]lib.Route, 0)
	peerConfigs := make([]wgtypes.PeerConfig, 0)
	self, err := params.mesh.GetNode(m.meshManager.GetPublicKey().String())

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

			cfg, err := m.convertMeshNode(convertMeshNodeParams{
				node:          n,
				mesh:          params.mesh,
				device:        params.dev,
				peerToClients: peerToClients,
				routes:        params.routes,
			})

			if err != nil {
				return nil, err
			}

			if NodeEquals(self, peer) {
				peerConfigs = append(peerConfigs, *cfg)
			}

			installedRoutes = append(installedRoutes, m.getRoutesToInstall(cfg, params.mesh, n)...)
		}
	}

	for _, n := range params.peers {
		if NodeEquals(n, self) {
			continue
		}

		peer, err := m.convertMeshNode(convertMeshNodeParams{
			node:          n,
			mesh:          params.mesh,
			peerToClients: peerToClients,
			routes:        params.routes,
			device:        params.dev,
		})

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

// updateWgConf: update the WireGuard configuration
func (m *WgMeshConfigApplyer) updateWgConf(mesh MeshProvider, routes map[string][]routeNode) error {
	snap, err := mesh.GetMesh()

	if err != nil {
		return err
	}

	nodes := lib.MapValues(snap.GetNodes())
	dev, err := mesh.GetDevice()

	if err != nil {
		return err
	}

	slices.SortFunc(nodes, func(a, b MeshNode) int {
		return strings.Compare(string(a.GetType()), string(b.GetType()))
	})

	peers := lib.Filter(nodes, func(mn MeshNode) bool {
		return mn.GetType() == conf.PEER_ROLE
	})

	clients := lib.Filter(nodes, func(mn MeshNode) bool {
		return mn.GetType() == conf.CLIENT_ROLE
	})

	self, err := mesh.GetNode(m.meshManager.GetPublicKey().String())

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

// getAllRoutes: works out all the routes to install out of all the routes in the
// set of networks the node is a part of
func (m *WgMeshConfigApplyer) getAllRoutes() (map[string][]routeNode, error) {
	allRoutes := make(map[string][]routeNode)

	for _, mesh := range m.meshManager.GetMeshes() {
		routes, err := m.getRoutes(mesh)

		if err != nil {
			return nil, err
		}

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

	return allRoutes, nil
}

// ApplyConfig: apply the WireGuard configuration
func (m *WgMeshConfigApplyer) ApplyConfig() error {
	allRoutes, err := m.getAllRoutes()

	if err != nil {
		return err
	}

	for _, mesh := range m.meshManager.GetMeshes() {
		err := m.updateWgConf(mesh, allRoutes)

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *WgMeshConfigApplyer) SetMeshManager(manager MeshManager) {
	m.meshManager = manager
}

func NewWgMeshConfigApplyer() MeshConfigApplyer {
	return &WgMeshConfigApplyer{
		routeInstaller: route.NewRouteInstaller(),
		hashFunc: func(mn MeshNode) int {
			pubKey, _ := mn.GetPublicKey()
			return lib.HashString(pubKey.String())
		},
	}
}
