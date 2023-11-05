package wg

type WgInterfaceManipulatorStub struct{}

func (i *WgInterfaceManipulatorStub) CreateInterface(params *CreateInterfaceParams) error {
	return nil
}

func (i *WgInterfaceManipulatorStub) EnableInterface(ifName string, ip string) error {
	return nil
}
