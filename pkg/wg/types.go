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
	// Enable interface enables the given interface with
	// the IP. It overrides the IP at the interface
	EnableInterface(ifName string, ip string) error
}
