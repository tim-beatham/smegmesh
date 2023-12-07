package wg

import (
	"crypto"
	"crypto/rand"
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type WgInterfaceManipulatorImpl struct {
	client *wgctrl.Client
}

const hashLength = 6

// CreateInterface creates a WireGuard interface
func (m *WgInterfaceManipulatorImpl) CreateInterface(port int, privKey *wgtypes.Key) (string, error) {
	rtnl, err := lib.NewRtNetlinkConfig()

	if err != nil {
		return "", fmt.Errorf("failed to access link: %w", err)
	}
	defer rtnl.Close()

	randomBuf := make([]byte, 32)
	_, err = rand.Read(randomBuf)

	if err != nil {
		return "", err
	}

	md5 := crypto.MD5.New().Sum(randomBuf)
	md5Str := fmt.Sprintf("wg%x", md5)[:hashLength]

	err = rtnl.CreateLink(md5Str)

	if err != nil {
		return "", fmt.Errorf("failed to create link: %w", err)
	}

	var cfg wgtypes.Config = wgtypes.Config{
		PrivateKey: privKey,
		ListenPort: &port,
	}

	err = m.client.ConfigureDevice(md5Str, cfg)

	if err != nil {
		m.RemoveInterface(md5Str)
		return "", fmt.Errorf("failed to configure dev: %w", err)
	}

	logging.Log.WriteInfof("ip link set up dev %s type wireguard", md5Str)
	return md5Str, nil
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
