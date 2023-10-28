package wg

import (
	"fmt"
	"net"
	"os/exec"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// createInterface uses ip link to create an interface. If the interface exists
// it returns an error
func createInterface(ifName string) error {
	_, err := net.InterfaceByName(ifName)

	if err == nil {
		return &WgError{msg: fmt.Sprintf("Interface %s already exists", ifName)}
	}

	// Check if the interface exists
	cmd := exec.Command("/usr/bin/ip", "link", "add", "dev", ifName, "type", "wireguard")

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

type WgInterfaceManipulatorImpl struct {
	client *wgctrl.Client
}

func (m *WgInterfaceManipulatorImpl) CreateInterface(params *CreateInterfaceParams) error {
	err := createInterface(params.IfName)

	if err != nil {
		return err
	}

	privateKey, err := wgtypes.GeneratePrivateKey()

	if err != nil {
		return err
	}

	var cfg wgtypes.Config = wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &params.Port,
	}

	m.client.ConfigureDevice(params.IfName, cfg)
	return nil
}

// flushInterface flushes the specified interface
func flushInterface(ifName string) error {
	_, err := net.InterfaceByName(ifName)

	if err != nil {
		return &WgError{msg: fmt.Sprintf("Interface %s does not exist cannot flush", ifName)}
	}

	cmd := exec.Command("/usr/bin/ip", "addr", "flush", "dev", ifName)

	if err := cmd.Run(); err != nil {
		logging.Log.WriteErrorf(fmt.Sprintf("%s error flushing interface %s", err.Error(), ifName))
		return &WgError{msg: fmt.Sprintf("Failed to flush interface %s", ifName)}
	}

	return nil
}

// EnableInterface flushes the interface and sets the ip address of the
// interface
func (m *WgInterfaceManipulatorImpl) EnableInterface(ifName string, ip string) error {
	err := flushInterface(ifName)

	if err != nil {
		return err
	}

	hostIp, _, err := net.ParseCIDR(ip)

	if err != nil {
		return err
	}

	cmd := exec.Command("/usr/bin/ip", "addr", "add", hostIp.String()+"/64", "dev", ifName)

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func NewWgInterfaceManipulator(client *wgctrl.Client) WgInterfaceManipulator {
	return &WgInterfaceManipulatorImpl{client: client}
}
