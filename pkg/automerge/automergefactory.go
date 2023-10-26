package crdt

import "github.com/tim-beatham/wgmesh/pkg/mesh"

type CrdtProviderFactory struct{}

func (f *CrdtProviderFactory) CreateMesh(params *mesh.MeshProviderFactoryParams) (mesh.MeshProvider, error) {
	return NewCrdtNodeManager(params.MeshId, params.DevName, params.Port,
		*params.Conf, params.Client)
}
