package meshtypes

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

/*
 * Represents a WireGuard mesh.
 */
type WgMesh struct {
	SharedKey wgtypes.Key
}

/*
 * Create a new WireGuard mesh with a new pre-shared key.
 */
func NewWgMesh() (*WgMesh, error) {

	key, err := wgtypes.GenerateKey()

	if err != nil {
		return nil, err
	}

	wgMesh := new(WgMesh)
	wgMesh.SharedKey = key
	return wgMesh, nil
}
