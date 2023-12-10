package crdt

import (
	"fmt"
	"hash/fnv"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

type TwoPhaseMapFactory struct {
	Config *conf.DaemonConfiguration
}

func (f *TwoPhaseMapFactory) CreateMesh(params *mesh.MeshProviderFactoryParams) (mesh.MeshProvider, error) {
	return &TwoPhaseStoreMeshManager{
		MeshId:     params.MeshId,
		IfName:     params.DevName,
		Client:     params.Client,
		conf:       params.Conf,
		daemonConf: params.DaemonConf,
		store: NewTwoPhaseMap[string, MeshNode](params.NodeID, func(s string) uint64 {
			h := fnv.New64a()
			h.Write([]byte(s))
			return h.Sum64()
		}, uint64(3*f.Config.KeepAliveTime)),
	}, nil
}

type MeshNodeFactory struct {
	Config conf.DaemonConfiguration
}

func (f *MeshNodeFactory) Build(params *mesh.MeshNodeFactoryParams) mesh.MeshNode {
	hostName := f.getAddress(params)

	grpcEndpoint := fmt.Sprintf("%s:%d", hostName, f.Config.GrpcPort)
	wgEndpoint := fmt.Sprintf("%s:%d", hostName, params.WgPort)

	if *params.MeshConfig.Role == conf.CLIENT_ROLE {
		grpcEndpoint = "-"
		wgEndpoint = "-"
	}

	return &MeshNode{
		HostEndpoint: grpcEndpoint,
		PublicKey:    params.PublicKey.String(),
		WgEndpoint:   wgEndpoint,
		WgHost:       fmt.Sprintf("%s/128", params.NodeIP.String()),
		Routes:       make(map[string]Route),
		Description:  "",
		Alias:        "",
		Type:         string(*params.MeshConfig.Role),
	}
}

// getAddress returns the routable address of the machine.
func (f *MeshNodeFactory) getAddress(params *mesh.MeshNodeFactoryParams) string {
	var hostName string = ""

	if params.Endpoint != "" {
		hostName = params.Endpoint
	} else if params.MeshConfig.Endpoint != nil && len(*params.MeshConfig.Endpoint) != 0 {
		hostName = *params.MeshConfig.Endpoint
	} else {
		ipFunc := lib.GetPublicIP

		if *params.MeshConfig.IPDiscovery == conf.DNS_IP_DISCOVERY {
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
