package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmespath/go-jmespath"
	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
)

// Querier queries a data store for the given data
// and returns data in the corresponding encoding
type Querier interface {
	Query(meshId string, queryParams string) ([]byte, error)
}

// JmesQuerier: queries the datstore in JMESPath syntax
type JmesQuerier struct {
	manager mesh.MeshManager
}

// QueryError: query error if something went wrong
type QueryError struct {
	msg string
}

// QuerRoute: represents a route in the query
type QueryRoute struct {
	Destination string `json:"destination"`
	HopCount    int    `json:"hopCount"`
	Path        string `json:"path"`
}

// QueryNode: represents a single node in the query
type QueryNode struct {
	HostEndpoint string            `json:"hostEndpoint"`
	PublicKey    string            `json:"publicKey"`
	WgEndpoint   string            `json:"wgEndpoint"`
	WgHost       string            `json:"wgHost"`
	Timestamp    int64             `json:"timestamp"`
	Description  string            `json:"description"`
	Routes       []QueryRoute      `json:"routes"`
	Alias        string            `json:"alias"`
	Services     map[string]string `json:"services"`
	Type         conf.NodeType     `json:"type"`
}

func (m *QueryError) Error() string {
	return m.msg
}

// Query: queries the the datastore at the given meshid
func (j *JmesQuerier) Query(meshId, queryParams string) ([]byte, error) {
	mesh, ok := j.manager.GetMeshes()[meshId]

	if !ok {
		return nil, &QueryError{msg: fmt.Sprintf("%s does not exist", meshId)}
	}

	snapshot, err := mesh.GetMesh()

	if err != nil {
		return nil, err
	}

	nodes := lib.Map(lib.MapValues(snapshot.GetNodes()), MeshNodeToQueryNode)

	result, err := jmespath.Search(queryParams, nodes)

	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(result)
	return bytes, err
}

// MeshNodeToQuerynode: convert the mesh node into a query abstraction
func MeshNodeToQueryNode(node mesh.MeshNode) *QueryNode {
	queryNode := new(QueryNode)
	queryNode.HostEndpoint = node.GetHostEndpoint()
	pubKey, _ := node.GetPublicKey()

	queryNode.PublicKey = pubKey.String()

	queryNode.WgEndpoint = node.GetWgEndpoint()
	queryNode.WgHost = node.GetWgHost().String()

	queryNode.Timestamp = node.GetTimeStamp()
	queryNode.Routes = lib.Map(node.GetRoutes(), func(r mesh.Route) QueryRoute {
		return QueryRoute{
			Destination: r.GetDestination().String(),
			HopCount:    r.GetHopCount(),
			Path:        strings.Join(r.GetPath(), ","),
		}
	})
	queryNode.Description = node.GetDescription()
	queryNode.Alias = node.GetAlias()
	queryNode.Services = node.GetServices()
	queryNode.Type = node.GetType()

	return queryNode
}

func NewJmesQuerier(manager mesh.MeshManager) Querier {
	return &JmesQuerier{manager: manager}
}
