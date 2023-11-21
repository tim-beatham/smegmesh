package wg

type WgInterfaceManipulatorStub struct{}

func (i *WgInterfaceManipulatorStub) CreateInterface(port int) (string, error) {
	return "", nil
}

func (i *WgInterfaceManipulatorStub) AddAddress(ifName string, addr string) error {
	return nil
}

func (i *WgInterfaceManipulatorStub) RemoveInterface(ifName string) error {
	return nil
}
