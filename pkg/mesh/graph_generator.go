package mesh

import (
	"errors"
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/graph"
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

type MeshGraphConverter interface {
	// convert the mesh to textual form
	Generate(meshId string) (string, error)
}

type MeshDOTConverter struct {
	manager *MeshManger
}

func (c *MeshDOTConverter) Generate(meshId string) (string, error) {
	mesh := c.manager.GetMesh(meshId)

	if mesh == nil {
		return "", errors.New("mesh does not exist")
	}

	g := graph.NewGraph(meshId, graph.GRAPH)

	snapshot, err := mesh.GetCrdt()

	if err != nil {
		return "", err
	}

	for _, node := range snapshot.Nodes {
		g.AddNode(node.GetEscapedIP())
	}

	nodes := lib.MapValues(snapshot.Nodes)

	for i, node1 := range nodes[:len(nodes)-1] {
		if mesh.HasFailed(node1.HostEndpoint) {
			continue
		}

		for _, node2 := range nodes[i+1:] {
			if node1.WgEndpoint == node2.WgEndpoint || mesh.HasFailed(node2.HostEndpoint) {
				continue
			}

			g.AddEdge(fmt.Sprintf("%s to %s", node1.GetEscapedIP(), node2.GetEscapedIP()), node1.GetEscapedIP(), node2.GetEscapedIP())
		}
	}

	return g.GetDOT()
}

func NewMeshDotConverter(m *MeshManger) MeshGraphConverter {
	return &MeshDOTConverter{manager: m}
}
