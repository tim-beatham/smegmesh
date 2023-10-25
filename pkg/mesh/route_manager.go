package mesh

import (
	"net"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/route"
)

type RouteManager interface {
	UpdateRoutes() error
	ApplyWg(mesh *crdt.CrdtNodeManager) error
}

type RouteManagerImpl struct {
	meshManager    *MeshManger
	routeInstaller route.RouteInstaller
}

func (r *RouteManagerImpl) UpdateRoutes() error {
	meshes := r.meshManager.Meshes
	ulaBuilder := new(ip.ULABuilder)

	for _, mesh1 := range meshes {
		for _, mesh2 := range meshes {
			if mesh1 == mesh2 {
				continue
			}

			ipNet, err := ulaBuilder.GetIPNet(mesh2.MeshId)

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
				return err
			}

			mesh1.AddRoutes(ipNet.String())
		}
	}

	return nil
}

func (r *RouteManagerImpl) ApplyWg(mesh *crdt.CrdtNodeManager) error {
	snapshot, err := mesh.GetCrdt()

	if err != nil {
		return err
	}

	for _, node := range snapshot.Nodes {
		if node.HostEndpoint == r.meshManager.HostEndpoint {
			continue
		}

		for route, _ := range node.Routes {
			_, netIP, err := net.ParseCIDR(route)

			if err != nil {
				return err
			}

			err = r.routeInstaller.InstallRoutes(mesh.IfName, netIP)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func NewRouteManager(m *MeshManger) RouteManager {
	return &RouteManagerImpl{meshManager: m, routeInstaller: route.NewRouteInstaller()}
}
