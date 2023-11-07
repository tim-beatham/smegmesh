package wg

import (
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type WgInterfaceManipulatorImpl struct {
	client *wgctrl.Client
}

// CreateInterface creates a WireGuard interface
func (m *WgInterfaceManipulatorImpl) CreateInterface(params *CreateInterfaceParams) error {
	rtnl, err := lib.NewRtNetlinkConfig()

	if err != nil {
		return fmt.Errorf("failed to access link: %w", err)
	}
	defer rtnl.Close()

	err = rtnl.CreateLink(params.IfName)

	if err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	privateKey, err := wgtypes.GeneratePrivateKey()

	if err != nil {
		return fmt.Errorf("failed to create private key: %w", err)
	}

	var cfg wgtypes.Config = wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &params.Port,
	}

	err = m.client.ConfigureDevice(params.IfName, cfg)

	if err != nil {
		return fmt.Errorf("failed to configure dev: %w", err)
	}

	logging.Log.WriteInfof("ip link set up dev %s type wireguard", params.IfName)
	return nil
}

// Add an address to the given interface
func (m *WgInterfaceManipulatorImpl) AddAddress(ifName string, addr string) error {
	rtnl, err := lib.NewRtNetlinkConfig()

	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	defer rtnl.Close()

	err = rtnl.AddAddress(ifName, addr)

	if err != nil {
		err = fmt.Errorf("failed to add address: %w", err)
	}

	return err
}

// RemoveInterface implements WgInterfaceManipulator.
func (*WgInterfaceManipulatorImpl) RemoveInterface(ifName string) error {
	rtnl, err := lib.NewRtNetlinkConfig()

	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	defer rtnl.Close()

	return rtnl.DeleteLink(ifName)
}

func NewWgInterfaceManipulator(client *wgctrl.Client) WgInterfaceManipulator {
	return &WgInterfaceManipulatorImpl{client: client}
}
