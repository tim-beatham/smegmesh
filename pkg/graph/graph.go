// Graph allows the definition of a DOT graph in golang
package graph

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/tim-beatham/wgmesh/pkg/lib"
)

type GraphType string
type Shape string

const (
	GRAPH   GraphType = "graph"
	DIGRAPH           = "digraph"
)

const (
	CIRCLE  Shape = "circle"
	STAR    Shape = "star"
	HEXAGON Shape = "hexagon"
)

type Graph struct {
	Type  GraphType
	Label string
	nodes map[string]*Node
	edges []Edge
}

type Node struct {
	Name  string
	Shape Shape
}

type Edge interface {
	Dottable
}

type DirectedEdge struct {
	Label string
	From  *Node
	To    *Node
}

type UndirectedEdge struct {
	Label string
	From  *Node
	To    *Node
}

// Dottable means an implementer can convert the struct to DOT representation
type Dottable interface {
	GetDOT() (string, error)
}

func NewGraph(label string, graphType GraphType) *Graph {
	return &Graph{Type: graphType, Label: label, nodes: make(map[string]*Node), edges: make([]Edge, 0)}
}

// PutNode: puts a node in the graph
func (g *Graph) PutNode(label string, shape Shape) error {
	_, exists := g.nodes[label]

	if exists {
		// If exists no need to add the ndoe
		return nil
	}

	g.nodes[label] = &Node{Name: label, Shape: shape}
	return nil
}

func writeContituents[D Dottable](result *strings.Builder, elements ...D) error {
	for _, node := range elements {
		dot, err := node.GetDOT()

		if err != nil {
			return err
		}

		_, err = result.WriteString(dot)

		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) GetDOT() (string, error) {
	var result strings.Builder

	_, err := result.WriteString(fmt.Sprintf("%s {\n", g.Type))

	if err != nil {
		return "", err
	}

	_, err = result.WriteString("node [colorscheme=set312];\n")

	if err != nil {
		return "", err
	}

	nodes := lib.MapValues(g.nodes)

	err = writeContituents(&result, nodes...)

	if err != nil {
		return "", err
	}

	err = writeContituents(&result, g.edges...)

	if err != nil {
		return "", err
	}

	_, err = result.WriteString("}")

	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (g *Graph) constructEdge(label string, from *Node, to *Node) Edge {
	switch g.Type {
	case DIGRAPH:
		return &DirectedEdge{Label: label, From: from, To: to}
	default:
		return &UndirectedEdge{Label: label, From: from, To: to}
	}
}

// AddEdge: adds an edge between two nodes in the graph
func (g *Graph) AddEdge(label string, from string, to string) error {
	fromNode, exists := g.nodes[from]

	if !exists {
		return errors.New(fmt.Sprintf("Node %s does not exist", from))
	}

	toNode, exists := g.nodes[to]

	if !exists {
		return errors.New(fmt.Sprintf("Node %s does not exist", to))
	}

	g.edges = append(g.edges, g.constructEdge(label, fromNode, toNode))
	return nil
}

const numColours = 12

func (n *Node) hash() int {
	h := fnv.New32a()
	h.Write([]byte(n.Name))
	return (int(h.Sum32()) % numColours) + 1
}

func (n *Node) GetDOT() (string, error) {
	return fmt.Sprintf("node[shape=%s, style=\"filled\", fillcolor=%d] %s;\n",
		n.Shape, n.hash(), n.Name), nil
}

func (e *DirectedEdge) GetDOT() (string, error) {
	return fmt.Sprintf("%s -> %s;\n", e.From.Name, e.To.Name), nil
}

func (e *UndirectedEdge) GetDOT() (string, error) {
	return fmt.Sprintf("%s -- %s;\n", e.From.Name, e.To.Name), nil
}
