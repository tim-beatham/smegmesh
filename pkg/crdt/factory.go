package crdt

import (
	"fmt"
	"hash/fnv"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

type TwoPhaseMapFactory struct{}

func (f *TwoPhaseMapFactory) CreateMesh(params *mesh.MeshProviderFactoryParams) (mesh.MeshProvider, error) {
	return &TwoPhaseStoreMeshManager{
		MeshId: params.MeshId,
		IfName: params.DevName,
		Client: params.Client,
		conf:   params.Conf,
		store: NewTwoPhaseMap[string, MeshNode](params.NodeID, func(s string) uint64 {
			h := fnv.New64a()
			h.Write([]byte(s))
			return h.Sum64()
		}, uint64(3*params.Conf.KeepAliveTime)),
	}, nil
}

type MeshNodeFactory struct {
	Config conf.WgMeshConfiguration
}

func (f *MeshNodeFactory) Build(params *mesh.MeshNodeFactoryParams) mesh.MeshNode {
	hostName := f.getAddress(params)

	grpcEndpoint := fmt.Sprintf("%s:%s", hostName, f.Config.GrpcPort)

	if f.Config.Role == conf.CLIENT_ROLE {
		grpcEndpoint = "-"
	}

	return &MeshNode{
		HostEndpoint: grpcEndpoint,
		PublicKey:    params.PublicKey.String(),
		WgEndpoint:   fmt.Sprintf("%s:%d", hostName, params.WgPort),
		WgHost:       fmt.Sprintf("%s/128", params.NodeIP.String()),
		Routes:       make(map[string]Route),
		Description:  "",
		Alias:        "",
		Type:         string(f.Config.Role),
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
		ipFunc := lib.GetPublicIP

		if f.Config.IPDiscovery == conf.DNS_IP_DISCOVERY {
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
