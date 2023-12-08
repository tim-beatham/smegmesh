package lib

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/jsimonetti/rtnetlink"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"golang.org/x/sys/unix"
)

type RtNetlinkConfig struct {
	conn *rtnetlink.Conn
}

func NewRtNetlinkConfig() (*RtNetlinkConfig, error) {
	conn, err := rtnetlink.Dial(nil)

	if err != nil {
		return nil, err
	}

	return &RtNetlinkConfig{conn: conn}, nil
}

const WIREGUARD_MTU = 1420

// Create a netlink interface if it does not exist. ifName is the name of the netlink interface
func (c *RtNetlinkConfig) CreateLink(ifName string) error {
	_, err := net.InterfaceByName(ifName)

	if err == nil {
		return fmt.Errorf("interface %s already exists", ifName)
	}

	err = c.conn.Link.New(&rtnetlink.LinkMessage{
		Family: unix.AF_UNSPEC,
		Flags:  unix.IFF_UP,
		Attributes: &rtnetlink.LinkAttributes{
			Name: ifName,
			Info: &rtnetlink.LinkInfo{Kind: "wireguard"},
			MTU:  uint32(WIREGUARD_MTU),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create wireguard interface: %w", err)
	}

	return nil
}

// Delete link delete the specified interface
func (c *RtNetlinkConfig) DeleteLink(ifName string) error {
	iface, err := net.InterfaceByName(ifName)

	if err != nil {
		return fmt.Errorf("failed to get interface %s %w", ifName, err)
	}

	err = c.conn.Link.Delete(uint32(iface.Index))

	if err != nil {
		return fmt.Errorf("failed to delete wg interface %w", err)
	}

	return nil
}

// AddAddress adds an address to the given interface.
func (c *RtNetlinkConfig) AddAddress(ifName string, address string) error {
	iface, err := net.InterfaceByName(ifName)

	if err != nil {
		return fmt.Errorf("failed to get interface %s error: %w", ifName, err)
	}

	addr, cidr, err := net.ParseCIDR(address)

	if err != nil {
		return fmt.Errorf("failed to parse CIDR %s error: %w", addr, err)
	}

	family := unix.AF_INET6

	ipv4 := cidr.IP.To4()

	if ipv4 != nil {
		family = unix.AF_INET
	}

	// Calculate the prefix length
	ones, _ := cidr.Mask.Size()

	// Calculate the broadcast IP
	// Only used when family is AF_INET
	var brd net.IP
	if ipv4 != nil {
		brd = make(net.IP, len(ipv4))
		binary.BigEndian.PutUint32(brd, binary.BigEndian.Uint32(ipv4)|^binary.BigEndian.Uint32(net.IP(cidr.Mask).To4()))
	}

	err = c.conn.Address.New(&rtnetlink.AddressMessage{
		Family:       uint8(family),
		PrefixLength: uint8(ones),
		Scope:        unix.RT_SCOPE_UNIVERSE,
		Index:        uint32(iface.Index),
		Attributes: &rtnetlink.AddressAttributes{
			Address:   addr,
			Local:     addr,
			Broadcast: brd,
		},
	})

	if err != nil {
		err = fmt.Errorf("failed to add address to link %w", err)
	}

	return err
}

// AddRoute: adds a route to the routing table.
// ifName is the intrface to add the route to
// gateway is the IP of the gateway device to hop to
// dst is the network prefix of the advertised destination
func (c *RtNetlinkConfig) AddRoute(ifName string, route Route) error {
	iface, err := net.InterfaceByName(ifName)

	if err != nil {
		return fmt.Errorf("failed accessing interface %s error %w", ifName, err)
	}

	gw := route.Gateway
	dst := route.Destination

	var family uint8 = unix.AF_INET6

	if dst.IP.To4() != nil {
		family = unix.AF_INET
	}

	attr := rtnetlink.RouteAttributes{
		Dst:      dst.IP,
		OutIface: uint32(iface.Index),
		Gateway:  gw,
	}

	ones, _ := dst.Mask.Size()

	err = c.conn.Route.Replace(&rtnetlink.RouteMessage{
		Family:     family,
		Table:      unix.RT_TABLE_MAIN,
		Protocol:   unix.RTPROT_BOOT,
		Scope:      unix.RT_SCOPE_LINK,
		Type:       unix.RTN_UNICAST,
		DstLength:  uint8(ones),
		Attributes: attr,
	})

	if err != nil {
		return fmt.Errorf("failed to add route %w", err)
	}

	return nil
}

// DeleteRoute deletes routes with the gateway and destination
func (c *RtNetlinkConfig) DeleteRoute(ifName string, route Route) error {
	iface, err := net.InterfaceByName(ifName)

	if err != nil {
		return fmt.Errorf("failed accessing interface %s error %w", ifName, err)
	}

	gw := route.Gateway
	dst := route.Destination

	var family uint8 = unix.AF_INET6

	if dst.IP.To4() != nil {
		family = unix.AF_INET
	}

	attr := rtnetlink.RouteAttributes{
		Dst:      dst.IP,
		OutIface: uint32(iface.Index),
		Gateway:  gw,
	}

	ones, _ := dst.Mask.Size()

	err = c.conn.Route.Delete(&rtnetlink.RouteMessage{
		Family:     family,
		Table:      unix.RT_TABLE_MAIN,
		Protocol:   unix.RTPROT_BOOT,
		Scope:      unix.RT_SCOPE_LINK,
		Type:       unix.RTN_UNICAST,
		DstLength:  uint8(ones),
		Attributes: attr,
	})

	if err != nil {
		return fmt.Errorf("failed to delete route %s", dst.IP.String())
	}

	return nil
}

type Route struct {
	Gateway     net.IP
	Destination net.IPNet
}

func (r1 Route) equal(r2 Route) bool {
	return r1.Gateway.String() == r2.Gateway.String() &&
		r1.Destination.String() == r2.Destination.String()
}

// DeleteRoutes deletes all routes not in exclude
func (c *RtNetlinkConfig) DeleteRoutes(ifName string, family uint8, exclude ...Route) error {
	routes, err := c.listRoutes(ifName, family)

	if err != nil {
		return err
	}

	ifRoutes := make([]Route, 0)

	for _, rtRoute := range routes {
		maskSize := 128

		if family == unix.AF_INET {
			maskSize = 32
		}

		cidr := net.CIDRMask(int(rtRoute.DstLength), maskSize)
		route := Route{
			Gateway:     rtRoute.Attributes.Gateway,
			Destination: net.IPNet{IP: rtRoute.Attributes.Dst, Mask: cidr},
		}

		ifRoutes = append(ifRoutes, route)
	}

	shouldExclude := func(r Route) bool {
		for _, route := range exclude {
			if route.equal(r) {
				return false
			}

			if family == unix.AF_INET && route.Destination.IP.To4() == nil {
				return false
			}

			if family == unix.AF_INET6 && route.Destination.IP.To16() == nil {
				return false
			}
		}
		return true
	}

	toDelete := Filter(ifRoutes, shouldExclude)

	for _, route := range toDelete {
		logging.Log.WriteInfof("Deleting route: %s", route.Destination.String())
		err := c.DeleteRoute(ifName, route)

		if err != nil {
			return err
		}
	}

	return nil
}

// listRoutes lists all routes on the interface
func (c *RtNetlinkConfig) listRoutes(ifName string, family uint8) ([]rtnetlink.RouteMessage, error) {
	iface, err := net.InterfaceByName(ifName)

	if err != nil {
		return nil, fmt.Errorf("failed accessing interface %s error %w", ifName, err)
	}

	routes, err := c.conn.Route.List()

	if err != nil {
		return nil, fmt.Errorf("failed to get route %w", err)
	}

	filterFunc := func(r rtnetlink.RouteMessage) bool {
		return r.Attributes.Gateway != nil && r.Attributes.OutIface == uint32(iface.Index)
	}

	routes = Filter(routes, filterFunc)
	return routes, nil
}

func (c *RtNetlinkConfig) Close() error {
	return c.conn.Close()
}
