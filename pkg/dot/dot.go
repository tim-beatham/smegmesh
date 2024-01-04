// Graph allows the definition of a DOT graph in golang
package graph

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/tim-beatham/smegmesh/pkg/lib"
)

type GraphType string
type Shape string

const (
	GRAPH   GraphType = "graph"
	DIGRAPH GraphType = "digraph"
)

const (
	CIRCLE        Shape = "circle"
	STAR          Shape = "star"
	HEXAGON       Shape = "hexagon"
	PARALLELOGRAM Shape = "parallelogram"
)

type Graph interface {
	Dottable
	GetType() GraphType
}

// Cluster: represents a subgraph in the graphs
type Cluster struct {
	Type  GraphType
	Name  string
	Label string
	nodes map[string]*Node
	edges map[string]Edge
}

// RootGraph: Represents the top level graph
type RootGraph struct {
	Type     GraphType
	Label    string
	nodes    map[string]*Node
	clusters map[string]*Cluster
	edges    map[string]Edge
}

// Node: represents a graphviz not
type Node struct {
	Name  string
	Label string
	Shape Shape
	Size  int
}

// Edge: represents an edge between adjacent nodes
type Edge interface {
	Dottable
}

// DirectEdge: contains a directed edge between any two nodes
type DirectedEdge struct {
	Name  string
	Label string
	From  string
	To    string
}

// UndirectedEdge: contains an undirected edge between any two
// nodes
type UndirectedEdge struct {
	Name  string
	Label string
	From  string
	To    string
}

// Dottable means an implementer can convert the struct to DOT representation
type Dottable interface {
	GetDOT() (string, error)
}

// PutNode: puts a node in the root graph
func (g *RootGraph) PutNode(name, label string, size int, shape Shape) error {
	_, exists := g.nodes[name]

	if exists {
		// If exists no need to add the ndoe
		return nil
	}

	g.nodes[name] = &Node{Name: name, Label: label, Size: size, Shape: shape}
	return nil
}

// PutCluster: puts a cluster in the root graph
func (g *RootGraph) PutCluster(graph *Cluster) {
	g.clusters[graph.Label] = graph
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

// GetDOT: convert the root graph into dot format
func (g *RootGraph) GetDOT() (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("%s {\n", g.Type))
	result.WriteString("node [colorscheme=set312];\n")
	result.WriteString("layout = fdp;\n")
	nodes := lib.MapValues(g.nodes)
	edges := lib.MapValues(g.edges)
	writeContituents(&result, nodes...)
	writeContituents(&result, edges...)

	for _, cluster := range g.clusters {
		clusterDOT, err := cluster.GetDOT()

		if err != nil {
			return "", err
		}

		result.WriteString(clusterDOT)
	}

	result.WriteString("}")
	return result.String(), nil
}

// GetType: get the graph type. DIRECTED|UNDIRECTED
func (r *RootGraph) GetType() GraphType {
	return r.Type
}

func constructEdge(graph Graph, name, label, from, to string) Edge {
	switch graph.GetType() {
	case DIGRAPH:
		return &DirectedEdge{Name: name, Label: label, From: from, To: to}
	default:
		return &UndirectedEdge{Name: name, Label: label, From: from, To: to}
	}
}

// AddEdge: adds an edge between two nodes in the root graph
func (g *RootGraph) AddEdge(name string, label string, from string, to string) error {
	g.edges[name] = constructEdge(g, name, label, from, to)
	return nil
}

const numColours = 12

func (n *Node) hash() int {
	h := fnv.New32a()
	h.Write([]byte(n.Name))
	return (int(h.Sum32()) % numColours) + 1
}

// GetDOT: convert the node into DOT format
func (n *Node) GetDOT() (string, error) {
	return fmt.Sprintf("node[label=\"%s\",shape=%s, style=\"filled\", fillcolor=%d, width=%d, height=%d, fixedsize=true] \"%s\";\n",
		n.Label, n.Shape, n.hash(), n.Size, n.Size, n.Name), nil
}

// GetDOT: Convert a directed edge into dot format
func (e *DirectedEdge) GetDOT() (string, error) {
	return fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", e.From, e.To, e.Label), nil
}

// GetDOT: convert an undirected edge into dot format
func (e *UndirectedEdge) GetDOT() (string, error) {
	return fmt.Sprintf("\"%s\" -- \"%s\" [label=\"%s\"];\n", e.From, e.To, e.Label), nil
}

// AddEdge: adds an edge between two nodes in the graph
func (g *Cluster) AddEdge(name string, label string, from string, to string) error {
	g.edges[name] = constructEdge(g, name, label, from, to)
	return nil
}

// PutNode: puts a node in the graph
func (g *Cluster) PutNode(name, label string, size int, shape Shape) error {
	_, exists := g.nodes[name]

	if exists {
		// If exists no need to add the ndoe
		return nil
	}

	g.nodes[name] = &Node{Name: name, Label: label, Shape: shape, Size: size}
	return nil
}

// GetDOT: convert the cluster into dot format
func (g *Cluster) GetDOT() (string, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("subgraph \"cluster%s\" {\n", g.Label))
	builder.WriteString(fmt.Sprintf("label = \"%s\"\n", g.Label))
	nodes := lib.MapValues(g.nodes)
	edges := lib.MapValues(g.edges)
	writeContituents(&builder, nodes...)
	writeContituents(&builder, edges...)

	builder.WriteString("}\n")
	return builder.String(), nil
}

// GetType: get the type of the subgraph (directed|undirected)
func (g *Cluster) GetType() GraphType {
	return g.Type
}

// NewSubGraph: instantiate a new subgraph
func NewSubGraph(name string, label string, graphType GraphType) *Cluster {
	return &Cluster{
		Label: name,
		Type:  graphType,
		Name:  name,
		nodes: make(map[string]*Node),
		edges: make(map[string]Edge),
	}
}

// NewGraph: create a new root graph
func NewGraph(label string, graphType GraphType) *RootGraph {
	return &RootGraph{
		Type: graphType, 
		Label: label, 
		clusters: map[string]*Cluster{}, 
		nodes: make(map[string]*Node), 
		edges: make(map[string]Edge)
	}
}