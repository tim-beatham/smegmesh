package automerge

import (
	"fmt"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
)

// CrdtProviderFactory: abstracts the instantiation of an automerge
// datastore
type CrdtProviderFactory struct{}

// CreateMesh: create a new mesh datastore
func (f *CrdtProviderFactory) CreateMesh(params *mesh.MeshProviderFactoryParams) (mesh.MeshProvider, error) {
	return NewCrdtNodeManager(&NewCrdtNodeMangerParams{
		MeshId:  params.MeshId,
		DevName: params.DevName,
		Conf:    params.Conf,
		Client:  params.Client,
	})
}

// MeshNodeFactory: abstracts the instnatiation of a node
type MeshNodeFactory struct {
	Config conf.DaemonConfiguration
}

// Build: builds the mesh node that represents the host machine to add
// to the  mesh
func (f *MeshNodeFactory) Build(params *mesh.MeshNodeFactoryParams) mesh.MeshNode {
	hostName := f.getAddress(params)

	grpcEndpoint := fmt.Sprintf("%s:%d", hostName, f.Config.GrpcPort)

	if *params.MeshConfig.Role == conf.CLIENT_ROLE {
		grpcEndpoint = "-"
	}

	return &MeshNodeCrdt{
		HostEndpoint: grpcEndpoint,
		PublicKey:    params.PublicKey.String(),
		WgEndpoint:   fmt.Sprintf("%s:%d", hostName, params.WgPort),
		WgHost:       fmt.Sprintf("%s/128", params.NodeIP.String()),
		// Always set the routes as empty.
		// Routes handled by external component
		Routes:      make(map[string]Route),
		Description: "",
		Alias:       "",
		Type:        string(*params.MeshConfig.Role),
	}
}

// getAddress: returns the routable address of the machine.
func (f *MeshNodeFactory) getAddress(params *mesh.MeshNodeFactoryParams) string {
	var hostName string = ""

	if params.Endpoint != "" {
		hostName = params.Endpoint
	} else if len(*params.MeshConfig.Endpoint) != 0 {
		hostName = *params.MeshConfig.Endpoint
	} else {
		ipFunc := lib.GetPublicIP

		if *params.MeshConfig.IPDiscovery == conf.OUTGOING_IP_DISCOVERY {
			ipFunc = lib.GetOutboundIP
		}

		ip, err := ipFunc()

		if err != nil {
			return ""
		}

		hostName = ip.String()
	}

	return hostName
}
