package wg

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

type WgInterfaceManipulatorStub struct{}

// CreateInterface creates a WireGuard interface
func (w *WgInterfaceManipulatorStub) CreateInterface(port int, privateKey *wgtypes.Key) (string, error) {
	return "aninterface", nil
}

// AddAddress adds an address to the given interface name
func (w *WgInterfaceManipulatorStub) AddAddress(ifName string, addr string) error {
	return nil
}

// RemoveInterface removes the specified interface
func (w *WgInterfaceManipulatorStub) RemoveInterface(ifName string) error {
	return nil
}
