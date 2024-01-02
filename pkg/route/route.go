package route

import (
	"github.com/tim-beatham/smegmesh/pkg/lib"
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

	err = rtnl.DeleteRoutes(devName, unix.AF_INET6, routes...)

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
