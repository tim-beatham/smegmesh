package wg

type WgInterfaceManipulatorStub struct{}

func (i *WgInterfaceManipulatorStub) CreateInterface(params *CreateInterfaceParams) error {
	return nil
}

func (i *WgInterfaceManipulatorStub) AddAddress(ifName string, addr string) error {
	return nil
}

func (i *WgInterfaceManipulatorStub) RemoveInterface(ifName string) error {
	return nil
}
