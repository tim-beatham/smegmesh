package wg

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

type WgError struct {
	msg string
}

func (m *WgError) Error() string {
	return m.msg
}

type WgInterfaceManipulator interface {
	// CreateInterface creates a WireGuard interface
	CreateInterface(port int, privateKey *wgtypes.Key) (string, error)
	// AddAddress adds an address to the given interface name
	AddAddress(ifName string, addr string) error
	// RemoveInterface removes the specified interface
	RemoveInterface(ifName string) error
}
