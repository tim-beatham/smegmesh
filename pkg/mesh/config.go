package mesh

import (
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
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
}

func (m *WgMeshConfigApplyer) convertMeshNode(node MeshNode, device *wgtypes.Device, peerToClients map[string][]net.IPNet) (*wgtypes.PeerConfig, error) {
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

	for _, route := range node.GetRoutes() {
		allowedips = append(allowedips, *route.GetDestination())
	}

	clients, ok := peerToClients[node.GetWgHost().String()]

	if ok {
		allowedips = append(allowedips, clients...)
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
	}

	return &peerConfig, nil
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

	routes := make([]lib.Route, 1)

	for _, n := range nodes {
		if NodeEquals(n, self) {
			continue
		}

		for _, route := range n.GetRoutes() {

			routes = append(routes, lib.Route{
				Gateway:     n.GetWgHost().IP,
				Destination: *route.GetDestination(),
			})
		}

		if n.GetType() == conf.CLIENT_ROLE && len(peers) > 0 && self.GetType() == conf.CLIENT_ROLE {
			peer := lib.ConsistentHash(peers, n, func(mn MeshNode) int {
				return lib.HashString(mn.GetWgHost().String())
			})

			clients, ok := peerToClients[peer.GetWgHost().String()]

			if !ok {
				clients = make([]net.IPNet, 0)
				peerToClients[peer.GetWgHost().String()] = clients
			}

			peerToClients[peer.GetWgHost().String()] = append(clients, *n.GetWgHost())
			continue
		}

		dev, _ := mesh.GetDevice()

		peer, err := m.convertMeshNode(n, dev, peerToClients)

		if err != nil {
			return err
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

	err = m.routeInstaller.InstallRoutes(dev.Name, routes...)

	if err != nil {
		return err
	}

	return m.meshManager.GetClient().ConfigureDevice(dev.Name, cfg)
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
		ReplacePeers: true,
		Peers:        make([]wgtypes.PeerConfig, 0),
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
