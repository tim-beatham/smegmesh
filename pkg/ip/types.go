package ip

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// IPAllocator: abstracts the process of creating an IP address
type IPAllocator interface {
	GetIP(key wgtypes.Key, meshId string, collisionCount uint8) (net.IP, error)
}
