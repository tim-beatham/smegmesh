package mesh

import (
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/hosts"
)

type MeshAliasManager interface {
	AddAliases(nodes []MeshNode)
	RemoveAliases(node []MeshNode)
}

type AliasManager struct {
	hosts hosts.HostsManipulator
}

// AddAliases: on node update or change add aliases to the hosts file
func (a *AliasManager) AddAliases(nodes []MeshNode) {
	for _, node := range nodes {
		if node.GetAlias() != "" {
			a.hosts.AddAddr(hosts.HostsEntry{
				Alias: fmt.Sprintf("%s.smeg", node.GetAlias()),
				Ip:    node.GetWgHost().IP,
			})
		}
	}
}

// RemoveAliases: on node remove remove aliases from the hosts file
func (a *AliasManager) RemoveAliases(nodes []MeshNode) {
	for _, node := range nodes {
		if node.GetAlias() != "" {
			a.hosts.Remove(hosts.HostsEntry{
				Alias: fmt.Sprintf("%s.smeg", node.GetAlias()),
				Ip:    node.GetWgHost().IP,
			})
		}
	}
}

func NewAliasManager() MeshAliasManager {
	return &AliasManager{
		hosts: hosts.NewHostsManipulator(),
	}
}
