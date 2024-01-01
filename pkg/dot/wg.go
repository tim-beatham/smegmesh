package graph

import (
	"fmt"
	"slices"

	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
)

// MeshGraphConverter converts a mesh to a graph
type MeshGraphConverter interface {
	// convert the mesh to textual form
	Generate() (string, error)
}

type MeshDOTConverter struct {
	meshes       map[string][]ctrlserver.MeshNode
	destinations map[string]interface{}
}

func (c *MeshDOTConverter) Generate() (string, error) {
	g := NewGraph("Smegmesh", GRAPH)

	for meshId := range c.meshes {
		err := c.generateMesh(g, meshId)

		if err != nil {
			return "", err
		}
	}

	for mesh := range c.meshes {
		g.PutNode(mesh, mesh, 1, CIRCLE)
	}

	for destination := range c.destinations {
		g.PutNode(destination, destination, 1, HEXAGON)
	}

	return g.GetDOT()
}

func (c *MeshDOTConverter) generateMesh(g *RootGraph, meshId string) error {
	nodes := c.meshes[meshId]

	g.PutNode(meshId, meshId, 1, CIRCLE)

	for _, node := range nodes {
		c.graphNode(g, node, meshId)
	}

	for _, node := range nodes {
		g.AddEdge(fmt.Sprintf("%s to %s", node.PublicKey, meshId), "", node.PublicKey, meshId)
	}

	return nil
}

// graphNode: graphs a node within the mesh
func (c *MeshDOTConverter) graphNode(g *RootGraph, node ctrlserver.MeshNode, meshId string) {
	alias := node.Alias

	if alias == "" {
		alias = node.WgHost[1:len(node.WgHost)-20] + "\\n" + node.WgHost[len(node.WgHost)-20:len(node.WgHost)]
	}

	g.PutNode(node.PublicKey, alias, 2, CIRCLE)

	for _, route := range node.Routes {
		if len(route.Path) == 0 {
			g.AddEdge(route.Destination, "", node.PublicKey, route.Destination)
			continue
		}

		reversedPath := slices.Clone(route.Path)
		slices.Reverse(reversedPath)

		g.AddEdge(fmt.Sprintf("%s to %s", node.PublicKey, reversedPath[0]), "", node.PublicKey, reversedPath[0])

		for _, mesh := range route.Path {
			if _, ok := c.meshes[mesh]; !ok {
				c.destinations[mesh] = struct{}{}
			}
		}

		for index := range reversedPath[0 : len(reversedPath)-1] {
			routeID := fmt.Sprintf("%s to %s", reversedPath[index], reversedPath[index+1])
			g.AddEdge(routeID, "", reversedPath[index], reversedPath[index+1])
		}

		if route.Destination == "::/0" {
			c.destinations[route.Destination] = struct{}{}
			lastMesh := reversedPath[len(reversedPath)-1]
			routeID := fmt.Sprintf("%s to %s", lastMesh, route.Destination)
			g.AddEdge(routeID, "", lastMesh, route.Destination)
		}
	}

	for service := range node.Services {
		c.putService(g, service, meshId, node)
	}
}

// putService: construct a service node and a link between the nodes
func (c *MeshDOTConverter) putService(g *RootGraph, key, meshId string, node ctrlserver.MeshNode) {
	serviceID := fmt.Sprintf("%s%s%s", key, node.PublicKey, meshId)
	g.PutNode(serviceID, key, 1, PARALLELOGRAM)
	g.AddEdge(fmt.Sprintf("%s to %s", node.PublicKey, serviceID), "", node.PublicKey, serviceID)
}

func NewMeshGraphConverter(meshes map[string][]ctrlserver.MeshNode) MeshGraphConverter {
	return &MeshDOTConverter{
		meshes:       meshes,
		destinations: make(map[string]interface{}),
	}
}
