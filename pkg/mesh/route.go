package mesh

import (
	"fmt"
	"net"

	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"golang.org/x/sys/unix"
)

type RouteManager interface {
	UpdateRoutes() error
	RemoveRoutes(meshId string) error
}

type RouteManagerImpl struct {
	meshManager MeshManager
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

		routes, err := mesh1.GetRoutes(pubKey.String())

		if err != nil {
			return err
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

			err = mesh2.AddRoutes(NodeID(self), append(lib.MapValues(routes), &RouteStub{
				Destination: ipNet,
				HopCount:    0,
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

// AddRoute adds a route to the given interface
func (m *RouteManagerImpl) addRoute(ifName string, meshPrefix string, routes ...lib.Route) error {
	rtnl, err := lib.NewRtNetlinkConfig()

	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	defer rtnl.Close()

	// Delete any routes that may be vacant
	err = rtnl.DeleteRoutes(ifName, unix.AF_INET6, routes...)

	if err != nil {
		return err
	}

	for _, route := range routes {
		if route.Destination.String() == meshPrefix {
			continue
		}

		err = rtnl.AddRoute(ifName, route)

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *RouteManagerImpl) installRoute(ifName string, meshid string, node MeshNode) error {
	routeMapFunc := func(route string) lib.Route {
		_, cidr, _ := net.ParseCIDR(route)

		r := lib.Route{
			Destination: *cidr,
			Gateway:     node.GetWgHost().IP,
		}
		return r
	}

	ipBuilder := &ip.ULABuilder{}
	ipNet, err := ipBuilder.GetIPNet(meshid)

	if err != nil {
		return err
	}

	theRoutes := lib.Map(node.GetRoutes(), func(r Route) string {
		return r.GetDestination().String()
	})

	routes := lib.Map(append(theRoutes, ipNet.String()), routeMapFunc)
	return m.addRoute(ifName, ipNet.String(), routes...)
}

func (m *RouteManagerImpl) installRoutes(meshProvider MeshProvider) error {
	mesh, err := meshProvider.GetMesh()

	if err != nil {
		return err
	}

	dev, err := meshProvider.GetDevice()

	if err != nil {
		return err
	}

	self, err := m.meshManager.GetSelf(meshProvider.GetMeshId())

	if err != nil {
		return err
	}

	for _, node := range mesh.GetNodes() {
		if NodeEquals(self, node) {
			continue
		}

		err = m.installRoute(dev.Name, meshProvider.GetMeshId(), node)

		if err != nil {
			return err
		}
	}

	return nil
}

// InstallRoutes installs all routes to the RIB
func (r *RouteManagerImpl) InstallRoutes() error {
	for _, mesh := range r.meshManager.GetMeshes() {
		err := r.installRoutes(mesh)

		if err != nil {
			return err
		}
	}

	return nil
}

func NewRouteManager(m MeshManager) RouteManager {
	return &RouteManagerImpl{meshManager: m}
}
