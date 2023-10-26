package mesh

import (
	"errors"
	"fmt"

	"github.com/tim-beatham/wgmesh/pkg/graph"
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

// MeshGraphConverter converts a mesh to a graph
type MeshGraphConverter interface {
	// convert the mesh to textual form
	Generate(meshId string) (string, error)
}

type MeshDOTConverter struct {
	manager *MeshManager
}

func (c *MeshDOTConverter) Generate(meshId string) (string, error) {
	mesh := c.manager.GetMesh(meshId)

	if mesh == nil {
		return "", errors.New("mesh does not exist")
	}

	g := graph.NewGraph(meshId, graph.GRAPH)

	snapshot, err := mesh.GetMesh()

	if err != nil {
		return "", err
	}

	for _, node := range snapshot.GetNodes() {
		g.AddNode(fmt.Sprintf("\"%s\"", node.GetWgHost().IP.String()))
	}

	nodes := lib.MapValues(snapshot.GetNodes())

	for i, node1 := range nodes[:len(nodes)-1] {
		for _, node2 := range nodes[i+1:] {
			if node1.GetWgEndpoint() == node2.GetWgEndpoint() {
				continue
			}

			node1Id := fmt.Sprintf("\"%s\"", node1.GetWgHost().IP.String())
			node2Id := fmt.Sprintf("\"%s\"", node2.GetWgHost().IP.String())
			g.AddEdge(fmt.Sprintf("%s to %s", node1Id, node2Id), node1Id, node2Id)
		}
	}

	return g.GetDOT()
}

func NewMeshDotConverter(m *MeshManager) MeshGraphConverter {
	return &MeshDOTConverter{manager: m}
}
