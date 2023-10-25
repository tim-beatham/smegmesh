package ip

import (
	"crypto/sha1"
	"fmt"
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type ULABuilder struct{}

func getMeshPrefix(meshId string) [16]byte {
	var ulaPrefix [16]byte

	ulaPrefix[0] = 0xfd

	s := sha1.Sum([]byte(meshId))

	for i := 1; i < 7; i++ {
		ulaPrefix[i] = s[i-1]
	}

	ulaPrefix[7] = 1
	return ulaPrefix
}

func (u *ULABuilder) GetIPNet(meshId string) (*net.IPNet, error) {
	meshBytes := getMeshPrefix(meshId)
	var meshIP net.IP = meshBytes[:]

	ip := fmt.Sprintf("%s/%d", meshIP.String(), 64)
	_, net, err := net.ParseCIDR(ip)

	if err != nil {
		return nil, err
	}

	return net, nil
}

func (u *ULABuilder) GetIP(key wgtypes.Key, meshId string) (net.IP, error) {
	ulaPrefix := getMeshPrefix(meshId)

	c, err := NewCga(key, ulaPrefix)

	if err != nil {
		return nil, err
	}
	return c.GetIP(), nil
}
