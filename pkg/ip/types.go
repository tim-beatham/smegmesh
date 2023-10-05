package ip

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type IPAllocator interface {
	GetIP(key wgtypes.Key, meshId string) (net.IP, error)
}
