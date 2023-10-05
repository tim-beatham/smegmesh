package ip

import (
	"crypto/sha1"
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type ULABuilder struct{}

func getULAPrefix(meshId string) [8]byte {
	var ulaPrefix [8]byte

	ulaPrefix[0] = 0xfd

	s := sha1.Sum([]byte(meshId))

	for i := 1; i < 7; i++ {
		ulaPrefix[i] = s[i-1]
	}

	ulaPrefix[7] = 1
	return ulaPrefix
}

func (u *ULABuilder) GetIP(key wgtypes.Key, meshId string) (net.IP, error) {
	ulaPrefix := getULAPrefix(meshId)

	c, err := NewCga(key, ulaPrefix)

	if err != nil {
		return nil, err
	}
	return c.GetIP(), nil
}
