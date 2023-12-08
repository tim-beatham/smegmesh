package mesh

import (
	"net"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

type RouteManager interface {
	UpdateRoutes() error
	RemoveRoutes(meshId string) error
}

type RouteManagerImpl struct {
	meshManager MeshManager
	conf        *conf.WgMeshConfiguration
}

func (r *RouteManagerImpl) UpdateRoutes() error {
	meshes := r.meshManager.GetMeshes()
	ulaBuilder := new(ip.ULABuilder)

	for _, mesh1 := range meshes {
		self, err := r.meshManager.GetSelf(mesh1.GetMeshId())

		if err != nil {
			return err
		}

		pubKey, err := self.GetPublicKey()

		if err != nil {
			return err
		}

		routeMap, err := mesh1.GetRoutes(pubKey.String())

		if err != nil {
			return err
		}

		if r.conf.AdvertiseDefaultRoute {
			_, ipv6Default, _ := net.ParseCIDR("::/0")

			mesh1.AddRoutes(NodeID(self),
				&RouteStub{
					Destination: ipv6Default,
					HopCount:    0,
					Path:        make([]string, 0),
				})
		}

		for _, mesh2 := range meshes {
			if mesh1 == mesh2 {
				continue
			}

			ipNet, err := ulaBuilder.GetIPNet(mesh2.GetMeshId())

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
				return err
			}

			routes := lib.MapValues(routeMap)

			err = mesh2.AddRoutes(NodeID(self), append(routes, &RouteStub{
				Destination: ipNet,
				HopCount:    0,
				Path:        make([]string, 0),
			})...)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// removeRoutes: removes all meshes we are no longer a part of
func (r *RouteManagerImpl) RemoveRoutes(meshId string) error {
	ulaBuilder := new(ip.ULABuilder)
	meshes := r.meshManager.GetMeshes()

	ipNet, err := ulaBuilder.GetIPNet(meshId)

	if err != nil {
		return err
	}

	for _, mesh1 := range meshes {
		self, err := r.meshManager.GetSelf(meshId)

		if err != nil {
			return err
		}

		mesh1.RemoveRoutes(NodeID(self), ipNet.String())
	}
	return nil
}

func NewRouteManager(m MeshManager, conf *conf.WgMeshConfiguration) RouteManager {
	return &RouteManagerImpl{meshManager: m, conf: conf}
}
