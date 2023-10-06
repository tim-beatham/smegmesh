package wg

import (
	"net"
	"os/exec"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
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
		logging.ErrorLog.Println(err.Error())
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
