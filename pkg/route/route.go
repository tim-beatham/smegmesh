package route

import (
	"net"
	"os/exec"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

type RouteInstaller interface {
	InstallRoutes(devName string, routes ...*net.IPNet) error
}

type RouteInstallerImpl struct{}

// InstallRoutes: installs a route into the routing table
func (r *RouteInstallerImpl) InstallRoutes(devName string, routes ...*net.IPNet) error {
	for _, route := range routes {
		err := r.installRoute(devName, route)

		if err != nil {
			return err
		}
	}

	return nil
}

// installRoute: installs a route into the linux table
func (r *RouteInstallerImpl) installRoute(devName string, route *net.IPNet) error {
	// TODO: Find a library that automates this
	cmd := exec.Command("/usr/bin/ip", "-6", "route", "add", route.String(), "dev", devName)

	logging.Log.WriteInfof("%s %s", route.String(), devName)

	if msg, err := cmd.CombinedOutput(); err != nil {
		logging.Log.WriteErrorf(err.Error())
		logging.Log.WriteErrorf(string(msg))
		return err
	}

	return nil
}

func NewRouteInstaller() RouteInstaller {
	return &RouteInstallerImpl{}
}
