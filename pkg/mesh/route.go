package mesh

import (
	"net"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

type RouteManager interface {
	UpdateRoutes() error
}

type RouteManagerImpl struct {
	meshManager MeshManager
	conf        *conf.WgMeshConfiguration
}

func (r *RouteManagerImpl) UpdateRoutes() error {
	meshes := r.meshManager.GetMeshes()
	routes := make(map[string][]Route)

	for _, mesh1 := range meshes {
		self, err := r.meshManager.GetSelf(mesh1.GetMeshId())

		if err != nil {
			return err
		}

		if _, ok := routes[mesh1.GetMeshId()]; !ok {
			routes[mesh1.GetMeshId()] = make([]Route, 0)
		}

		routeMap, err := mesh1.GetRoutes(NodeID(self))

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
			routeValues, ok := routes[mesh2.GetMeshId()]

			if !ok {
				routeValues = make([]Route, 0)
			}

			if mesh1 == mesh2 {
				continue
			}

			mesh1IpNet, _ := (&ip.ULABuilder{}).GetIPNet(mesh1.GetMeshId())

			routeValues = append(routeValues, &RouteStub{
				Destination: mesh1IpNet,
				HopCount:    0,
				Path:        []string{mesh1.GetMeshId()},
			})

			routeValues = append(routeValues, lib.MapValues(routeMap)...)
			mesh2IpNet, _ := (&ip.ULABuilder{}).GetIPNet(mesh2.GetMeshId())
			routeValues = lib.Filter(routeValues, func(r Route) bool {
				pathNotMesh := func(s string) bool {
					return s == mesh2.GetMeshId()
				}

				// Ensure that the route does not see it's own IP
				return !r.GetDestination().IP.Equal(mesh2IpNet.IP) && !lib.Contains(r.GetPath()[1:], pathNotMesh)
			})

			routes[mesh2.GetMeshId()] = routeValues
		}
	}

	// Calculate the set different of each, working out routes to remove and to keep.
	for meshId, meshRoutes := range routes {
		mesh := r.meshManager.GetMesh(meshId)
		self, _ := r.meshManager.GetSelf(meshId)
		toRemove := make([]Route, 0)

		prevRoutes, _ := mesh.GetRoutes(NodeID(self))

		for _, route := range prevRoutes {
			if !lib.Contains(meshRoutes, func(r Route) bool {
				return RouteEquals(r, route)
			}) {
				toRemove = append(toRemove, route)
			}
		}

		mesh.RemoveRoutes(NodeID(self), toRemove...)
		mesh.AddRoutes(NodeID(self), meshRoutes...)
	}

	return nil
}

func NewRouteManager(m MeshManager, conf *conf.WgMeshConfiguration) RouteManager {
	return &RouteManagerImpl{meshManager: m, conf: conf}
}
