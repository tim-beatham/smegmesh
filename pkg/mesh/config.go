package mesh

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// MeshConfigApplyer abstracts applying the mesh configuration
type MeshConfigApplyer interface {
	ApplyConfig() error
}

// WgMeshConfigApplyer applies WireGuard configuration
type WgMeshConfigApplyer struct {
	meshManager *MeshManager
}

func convertMeshNode(node MeshNode) (*wgtypes.PeerConfig, error) {
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

	peerConfig := wgtypes.PeerConfig{
		PublicKey:  pubKey,
		Endpoint:   endpoint,
		AllowedIPs: allowedips,
	}

	return &peerConfig, nil
}

func (m *WgMeshConfigApplyer) updateWgConf(mesh MeshProvider) error {
	snap, err := mesh.GetMesh()

	if err != nil {
		return err
	}

	nodes := snap.GetNodes()
	peerConfigs := make([]wgtypes.PeerConfig, len(nodes))

	var count int = 0

	for _, n := range nodes {
		peer, err := convertMeshNode(n)

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

	return m.meshManager.Client.ConfigureDevice(dev.Name, cfg)
}

func (m *WgMeshConfigApplyer) ApplyConfig() error {
	for _, mesh := range m.meshManager.Meshes {
		err := m.updateWgConf(mesh)

		if err != nil {
			return err
		}
	}

	return nil
}

func NewWgMeshConfigApplyer(manager *MeshManager) MeshConfigApplyer {
	return &WgMeshConfigApplyer{meshManager: manager}
}
