package mesh

import (
	"github.com/tim-beatham/wgmesh/pkg/route"
)

type RouteManager interface {
	UpdateRoutes() error
	ApplyWg() error
}

type RouteManagerImpl struct {
	meshManager    *MeshManager
	routeInstaller route.RouteInstaller
}

func (r *RouteManagerImpl) UpdateRoutes() error {
	// // meshes := r.meshManager.Meshes
	// // ulaBuilder := new(ip.ULABuilder)

	// for _, mesh1 := range meshes {
	// 	for _, mesh2 := range meshes {
	// 		if mesh1 == mesh2 {
	// 			continue
	// 		}

	// 		ipNet, err := ulaBuilder.GetIPNet(mesh2.MeshId)

	// 		if err != nil {
	// 			logging.Log.WriteErrorf(err.Error())
	// 			return err
	// 		}

	// 		mesh1.AddRoutes(ipNet.String())
	// 	}
	// }

	return nil
}

func (r *RouteManagerImpl) ApplyWg() error {
	// snapshot, err := mesh.GetMesh()

	// if err != nil {
	// 	return err
	// }

	// for _, node := range snapshot.Nodes {
	// 	if node.HostEndpoint == r.meshManager.HostEndpoint {
	// 		continue
	// 	}

	// 	for route, _ := range node.Routes {
	// 		_, netIP, err := net.ParseCIDR(route)

	// 		if err != nil {
	// 			return err
	// 		}

	// 		err = r.routeInstaller.InstallRoutes(mesh.IfName, netIP)

	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	return nil
}

func NewRouteManager(m *MeshManager) RouteManager {
	return &RouteManagerImpl{meshManager: m, routeInstaller: route.NewRouteInstaller()}
}
