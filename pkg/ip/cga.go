package ip

/*
 * Use a WireGuard public key to generate a unique interface ID
 */

import (
	"crypto/rand"
	"crypto/sha1"
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	ModifierLength = 16
	ZeroLength     = 9
	hash2Length    = 57
	hash1Length    = 58
	Hash2Prefix    = 14
	Hash1Prefix    = 8
	InterfaceIdLen = 8
)

/*
 * Cga parameters used to generate an IPV6 interface ID
 */
type CgaParameters struct {
	Modifier       [ModifierLength]byte
	SubnetPrefix   [2 * InterfaceIdLen]byte
	CollisionCount uint8
	PublicKey      wgtypes.Key
	interfaceId    [2 * InterfaceIdLen]byte
	flag           byte
}

func NewCga(key wgtypes.Key, collisionCount uint8, subnetPrefix [2 * InterfaceIdLen]byte) (*CgaParameters, error) {
	var params CgaParameters

	_, err := rand.Read(params.Modifier[:])

	if err != nil {
		return nil, err
	}

	params.PublicKey = key
	params.SubnetPrefix = subnetPrefix
	params.CollisionCount = collisionCount
	return &params, nil
}

func (c *CgaParameters) generateHash2() []byte {
	var byteVal [hash2Length]byte

	for i := 0; i < ModifierLength; i++ {
		byteVal[i] = c.Modifier[i]
	}

	for i := 0; i < wgtypes.KeyLen; i++ {
		byteVal[ModifierLength+ZeroLength+i] = c.PublicKey[i]
	}

	hash := sha1.Sum(byteVal[:])

	return hash[:Hash2Prefix]
}

func (c *CgaParameters) generateHash1() []byte {
	var byteVal [hash1Length]byte

	for i := 0; i < ModifierLength; i++ {
		byteVal[i] = c.Modifier[i]
	}

	for i := 0; i < wgtypes.KeyLen; i++ {
		byteVal[ModifierLength+ZeroLength+i] = c.PublicKey[i]
	}

	byteVal[hash1Length-1] = c.CollisionCount

	hash := sha1.Sum(byteVal[:])
	return hash[:Hash1Prefix]
}

func clearBit(num, pos int) byte {
	mask := ^(1 << pos)
	result := num & mask

	return byte(result)
}

func (c *CgaParameters) generateInterface() []byte {
	hash1 := c.generateHash1()

	var interfaceId []byte = make([]byte, InterfaceIdLen)

	copy(interfaceId[:], hash1)

	interfaceId[0] = clearBit(int(interfaceId[0]), 6)
	interfaceId[0] = clearBit(int(interfaceId[1]), 7)

	return interfaceId
}

func (c *CgaParameters) GetIP() net.IP {
	if c.flag == 1 {
		return c.interfaceId[:]
	}

	bytes := c.generateInterface()

	for i := 0; i < InterfaceIdLen; i++ {
		c.interfaceId[i] = c.SubnetPrefix[i]
	}

	for i := InterfaceIdLen; i < 2*InterfaceIdLen; i++ {
		c.interfaceId[i] = bytes[i-8]
	}

	c.flag = 1
	return c.interfaceId[:]
}
