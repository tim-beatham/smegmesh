package mesh

import (
	"github.com/tim-beatham/wgmesh/pkg/ip"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/route"
)

type RouteManager interface {
	UpdateRoutes() error
}

type RouteManagerImpl struct {
	meshManager    MeshManager
	routeInstaller route.RouteInstaller
}

func (r *RouteManagerImpl) UpdateRoutes() error {
	meshes := r.meshManager.GetMeshes()
	ulaBuilder := new(ip.ULABuilder)

	for _, mesh1 := range meshes {
		for _, mesh2 := range meshes {
			if mesh1 == mesh2 {
				continue
			}

			ipNet, err := ulaBuilder.GetIPNet(mesh2.GetMeshId())

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
				return err
			}

			self, err := r.meshManager.GetSelf(mesh1.GetMeshId())

			if err != nil {
				return err
			}

			err = mesh1.AddRoutes(self.GetHostEndpoint(), ipNet.String())

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func NewRouteManager(m MeshManager) RouteManager {
	return &RouteManagerImpl{meshManager: m, routeInstaller: route.NewRouteInstaller()}
}
