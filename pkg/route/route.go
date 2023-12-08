package route

import (
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"golang.org/x/sys/unix"
)

type RouteInstaller interface {
	InstallRoutes(devName string, routes ...lib.Route) error
}

type RouteInstallerImpl struct{}

// InstallRoutes: installs a route into the routing table
func (r *RouteInstallerImpl) InstallRoutes(devName string, routes ...lib.Route) error {
	rtnl, err := lib.NewRtNetlinkConfig()

	if err != nil {
		return err
	}

	ip6Routes := lib.Filter(routes, func(r lib.Route) bool {
		return r.Destination.IP.To4() == nil
	})

	err = rtnl.DeleteRoutes(devName, unix.AF_INET6, ip6Routes...)

	if err != nil {
		return err
	}

	for _, route := range routes {
		err := rtnl.AddRoute(devName, route)

		if err != nil {
			return err
		}
	}

	return nil
}

func NewRouteInstaller() RouteInstaller {
	return &RouteInstallerImpl{}
}
