package mesh

type RouteManagerStub struct {
}

func (r *RouteManagerStub) UpdateRoutes() error {
	return nil
}

func (r *RouteManagerStub) InstallRoutes() error {
	return nil
}

func (r *RouteManagerStub) RemoveRoutes(meshId string) error {
	return nil
}
