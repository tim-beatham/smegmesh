package slaac

import (
	"crypto/sha1"

	"github.com/tim-beatham/wgmesh/pkg/cga"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type ULA struct {
	CGA cga.CgaParameters
}

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

func NewULA(key wgtypes.Key, meshId string) (*ULA, error) {
	ulaPrefix := getULAPrefix(meshId)

	c, err := cga.NewCga(key, ulaPrefix)

	if err != nil {
		return nil, err
	}

	return &ULA{CGA: *c}, nil
}
