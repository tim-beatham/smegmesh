package mesh

import "github.com/tim-beatham/wgmesh/pkg/hosts"

func AddAliases(meshid string, snapshot MeshSnapshot) {
	hosts := hosts.NewHostsManipulator(meshid)

	for _, node := range snapshot.GetNodes() {
		if node.GetAlias() != "" {
			hosts.AddAddr(node.GetWgHost().IP, node.GetAlias())
		}
	}

	hosts.Write()
}
