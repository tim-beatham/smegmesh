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
	manager MeshManager
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
		c.graphNode(g, node, meshId)
	}

	nodes := lib.MapValues(snapshot.GetNodes())

	for i, node1 := range nodes[:len(nodes)-1] {
		for _, node2 := range nodes[i+1:] {
			if node1.GetWgEndpoint() == node2.GetWgEndpoint() {
				continue
			}

			node1Id := fmt.Sprintf("\"%s\"", node1.GetIdentifier())
			node2Id := fmt.Sprintf("\"%s\"", node2.GetIdentifier())
			g.AddEdge(fmt.Sprintf("%s to %s", node1Id, node2Id), node1Id, node2Id)
		}
	}

	return g.GetDOT()
}

// graphNode: graphs a node within the mesh
func (c *MeshDOTConverter) graphNode(g *graph.Graph, node MeshNode, meshId string) {
	nodeId := fmt.Sprintf("\"%s\"", node.GetIdentifier())
	g.PutNode(nodeId, graph.CIRCLE)

	self, _ := c.manager.GetSelf(meshId)

	if NodeEquals(self, node) {
		return
	}

	for _, route := range node.GetRoutes() {
		routeId := fmt.Sprintf("\"%s\"", route)
		g.PutNode(routeId, graph.HEXAGON)
		g.AddEdge(fmt.Sprintf("%s to %s", nodeId, routeId), nodeId, routeId)
	}
}

func NewMeshDotConverter(m MeshManager) MeshGraphConverter {
	return &MeshDOTConverter{manager: m}
}
