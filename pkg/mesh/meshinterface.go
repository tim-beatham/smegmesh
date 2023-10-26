package mesh

import (
	"errors"

	"github.com/tim-beatham/wgmesh/pkg/wg"
)

// MeshInterfaces manipulates interfaces to do with meshes
type MeshInterface interface {
	EnableInterface(meshId string) error
}

type WgMeshInterface struct {
	manager *MeshManager
}

// EnableInterface enables the interface at the given endpoint
func (m *WgMeshInterface) EnableInterface(meshId string) error {
	mesh, ok := m.manager.Meshes[meshId]

	if !ok {
		return errors.New("the provided mesh does not exist")
	}

	dev, err := mesh.GetDevice()

	if err != nil {
		return err
	}

	self, err := m.manager.GetSelf(meshId)

	if err != nil {
		return err
	}

	return wg.EnableInterface(dev.Name, self.GetWgHost().String())
}

func NewWgMeshInterface(manager *MeshManager) MeshInterface {
	return &WgMeshInterface{manager: manager}
}
