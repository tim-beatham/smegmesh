package mesh

import (
	"net"

	"github.com/tim-beatham/smegmesh/pkg/ip"
	"github.com/tim-beatham/smegmesh/pkg/lib"
)

// RouteManager: manager that leaks routes between meshes
type RouteManager interface {
	// UpdateRoutes: leak all routes in each mesh
	UpdateRoutes() error
}

type RouteManagerImpl struct {
	meshManager MeshManager
}

func (r *RouteManagerImpl) UpdateRoutes() error {
	meshes := r.meshManager.GetMeshes()
	routes := make(map[string][]Route)

	for _, mesh := range meshes {
		// Make empty routes so that routes are retracted
		routes[mesh.GetMeshId()] = make([]Route, 0)
	}

	for _, mesh1 := range meshes {
		if !*mesh1.GetConfiguration().AdvertiseRoutes {
			continue
		}

		self, err := mesh1.GetNode(r.meshManager.GetPublicKey().String())

		if err != nil {
			return err
		}

		if _, ok := routes[mesh1.GetMeshId()]; !ok {
			routes[mesh1.GetMeshId()] = make([]Route, 0)
		}

		if *mesh1.GetConfiguration().AdvertiseDefaultRoute {
			_, ipv6Default, _ := net.ParseCIDR("::/0")

			defaultRoute := &RouteStub{
				Destination: ipv6Default,
				Path:        []string{mesh1.GetMeshId()},
			}

			mesh1.AddRoutes(NodeID(self), defaultRoute)
			routes[mesh1.GetMeshId()] = append(routes[mesh1.GetMeshId()], defaultRoute)
		}

		routeMap, err := mesh1.GetRoutes(NodeID(self))

		if err != nil {
			return err
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
				Path:        []string{mesh1.GetMeshId()},
			})

			routeValues = append(routeValues, lib.MapValues(routeMap)...)
			mesh2IpNet, _ := (&ip.ULABuilder{}).GetIPNet(mesh2.GetMeshId())
			routeValues = lib.Filter(routeValues, func(r Route) bool {
				pathNotMesh := func(s string) bool {
					return s == mesh2.GetMeshId()
				}

				// Remove any potential routing loops
				return !r.GetDestination().IP.Equal(mesh2IpNet.IP) &&
					!lib.Contains(r.GetPath()[1:], pathNotMesh)
			})

			routes[mesh2.GetMeshId()] = routeValues
		}
	}

	// Calculate the set different of each, working out routes to remove and to keep.
	for meshId, meshRoutes := range routes {
		mesh := meshes[meshId]

		self, err := mesh.GetNode(r.meshManager.GetPublicKey().String())

		if err != nil {
			return err
		}

		toRemove := make([]Route, 0)

		prevRoutes := self.GetRoutes()

		for _, route := range prevRoutes {
			if !lib.Contains(meshRoutes, func(r Route) bool {
				return RouteEqual(r, route)
			}) {
				toRemove = append(toRemove, route)
			}
		}

		mesh.RemoveRoutes(NodeID(self), toRemove...)
		mesh.AddRoutes(NodeID(self), meshRoutes...)
	}

	return nil
}

func NewRouteManager(m MeshManager) RouteManager {
	return &RouteManagerImpl{meshManager: m}
}
