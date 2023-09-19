package wg

import (
	"fmt"
	"net"
	"os/exec"

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
			fmt.Println(err.Error())
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

	wgListenPort := 5000
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
