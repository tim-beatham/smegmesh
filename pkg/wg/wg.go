package wg

import (
	"fmt"
	"net"
	"os/exec"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

/*
 * All WireGuard mesh interface called wgmesh
 */
func CreateInterface(ifName string) error {
	_, err := net.InterfaceByName(ifName)

	// Check if the interface exists
	if err != nil {
		cmd := exec.Command("/usr/bin/ip", "link", "add", "dev", ifName, "type", "wireguard")

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

/*
 * Create and configure a new WireGuard client
 */
func CreateClient(ifName string) (*wgctrl.Client, error) {
	err := CreateInterface(ifName)

	if err != nil {
		return nil, err
	}

	client, err := wgctrl.New()

	if err != nil {
		return nil, err
	}

	wgListenPort := 51820
	privateKey, err := wgtypes.GeneratePrivateKey()

	if err != nil {
		return nil, err
	}

	var cfg wgtypes.Config = wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &wgListenPort,
	}

	client.ConfigureDevice(ifName, cfg)
	return client, nil
}

func EnableInterface(ifName string, ip string) error {
	cmd := exec.Command("/usr/bin/ip", "link", "set", "up", "dev", ifName)

	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return err
	}

	hostIp, _, err := net.ParseCIDR(ip)

	if err != nil {
		return err
	}

	cmd = exec.Command("/usr/bin/ip", "addr", "add", hostIp.String()+"/64", "dev", "wgmesh")

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func convertMeshNode(node crdt.MeshNodeCrdt) (*wgtypes.PeerConfig, error) {
	peerEndpoint, err := net.ResolveUDPAddr("udp", node.WgEndpoint)

	if err != nil {
		return nil, err
	}

	peerPublic, err := wgtypes.ParseKey(node.PublicKey)

	if err != nil {
		return nil, err
	}

	allowedIps := make([]net.IPNet, 1)
	_, ipnet, err := net.ParseCIDR(node.WgHost)

	if err != nil {
		return nil, err
	}

	allowedIps[0] = *ipnet

	peerConfig := wgtypes.PeerConfig{
		PublicKey:  peerPublic,
		Endpoint:   peerEndpoint,
		AllowedIPs: allowedIps,
	}

	return &peerConfig, nil
}

func UpdateWgConf(devName string, nodes map[string]crdt.MeshNodeCrdt, client wgctrl.Client) error {
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
		Peers:        peerConfigs,
		ReplacePeers: true,
	}

	client.ConfigureDevice(devName, cfg)
	return nil
}
