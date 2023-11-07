package wg

type WgError struct {
	msg string
}

func (m *WgError) Error() string {
	return m.msg
}

type CreateInterfaceParams struct {
	IfName string
	Port   int
}

type WgInterfaceManipulator interface {
	// CreateInterface creates a WireGuard interface
	CreateInterface(params *CreateInterfaceParams) error
	// AddAddress adds an address to the given interface name
	AddAddress(ifName string, addr string) error
	// RemoveInterface removes the specified interface
	RemoveInterface(ifName string) error
}
