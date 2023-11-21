package mesh

import (
	"fmt"
	"hash/fnv"
	"net"
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

func (m *WgMeshConfigApplyer) convertMeshNode(node MeshNode) (*wgtypes.PeerConfig, error) {
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
		_, ipnet, _ := net.ParseCIDR(route)
		allowedips = append(allowedips, *ipnet)
	}

	keepAlive := time.Duration(m.config.KeepAliveTime) * time.Second

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

	rtnl, err := lib.NewRtNetlinkConfig()

	if err != nil {
		return err
	}

	for _, n := range nodes {
		if n.GetType() == conf.CLIENT_ROLE && len(peers) > 0 {
			a := fnv.New32a()
			a.Write([]byte(n.GetHostEndpoint()))
			sum := a.Sum32()

			responsiblePeer := peers[int(sum)%len(peers)]

			if responsiblePeer.GetHostEndpoint() != self.GetHostEndpoint() {
				dev, err := mesh.GetDevice()

				if err != nil {
					return err
				}

				rtnl.AddRoute(dev.Name, lib.Route{
					Gateway:     responsiblePeer.GetWgHost().IP,
					Destination: *n.GetWgHost(),
				})

				if err != nil {
					return err
				}

				continue
			}
		}

		peer, err := m.convertMeshNode(n)

		if err != nil {
			return err
		}

		peerConfigs[count] = *peer
		count++
	}

	cfg := wgtypes.Config{
		Peers: peerConfigs,
	}

	dev, err := mesh.GetDevice()

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
		Peers:        make([]wgtypes.PeerConfig, 1),
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
