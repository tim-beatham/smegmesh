package crdt

import (
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

type CrdtProviderFactory struct{}

func (f *CrdtProviderFactory) CreateMesh(params *mesh.MeshProviderFactoryParams) (mesh.MeshProvider, error) {
	return NewCrdtNodeManager(&NewCrdtNodeMangerParams{
		MeshId:  params.MeshId,
		DevName: params.DevName,
		Conf:    *params.Conf,
		Client:  params.Client,
	})
}

type MeshNodeFactory struct {
	Config conf.WgMeshConfiguration
}

// Build builds the mesh node that represents the host machine to add
// to the  mesh
func (f *MeshNodeFactory) Build(params *mesh.MeshNodeFactoryParams) mesh.MeshNode {
	hostName := f.getAddress(params)

	return &MeshNodeCrdt{
		HostEndpoint: fmt.Sprintf("%s:%s", hostName, f.Config.GrpcPort),
		PublicKey:    params.PublicKey.String(),
		WgEndpoint:   fmt.Sprintf("%s:%d", hostName, params.WgPort),
		WgHost:       fmt.Sprintf("%s/128", params.NodeIP.String()),
		// Always set the routes as empty.
		// Routes handled by external component
		Routes:      map[string]interface{}{},
		Description: "",
		Alias:       "",
		Type:        string(params.Role),
	}
}

// getAddress returns the routable address of the machine.
func (f *MeshNodeFactory) getAddress(params *mesh.MeshNodeFactoryParams) string {
	var hostName string = ""

	if params.Endpoint != "" {
		hostName = params.Endpoint
	} else if len(f.Config.Endpoint) != 0 {
		hostName = f.Config.Endpoint
	} else {
		hostName = lib.GetOutboundIP().String()
	}

	return hostName
}
