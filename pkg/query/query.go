package query

import (
	"encoding/json"
	"fmt"

	"github.com/jmespath/go-jmespath"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

// Querier queries a data store for the given data
// and returns data in the corresponding encoding
type Querier interface {
	Query(meshId string, queryParams string) ([]byte, error)
}

type JmesQuerier struct {
	manager mesh.MeshManager
}

type QueryError struct {
	msg string
}

type QueryNode struct {
	HostEndpoint string   `json:"hostEndpoint"`
	PublicKey    string   `json:"publicKey"`
	WgEndpoint   string   `json:"wgEndpoint"`
	WgHost       string   `json:"wgHost"`
	Timestamp    int64    `json:"timestmap"`
	Description  string   `json:"description"`
	Routes       []string `json:"routes"`
	Alias        string   `json:"alias"`
}

func (m *QueryError) Error() string {
	return m.msg
}

// Query: queries the data
func (j *JmesQuerier) Query(meshId, queryParams string) ([]byte, error) {
	mesh, ok := j.manager.GetMeshes()[meshId]

	if !ok {
		return nil, &QueryError{msg: fmt.Sprintf("%s does not exist", meshId)}
	}

	snapshot, err := mesh.GetMesh()

	if err != nil {
		return nil, err
	}

	nodes := lib.Map(lib.MapValues(snapshot.GetNodes()), meshNodeToQueryNode)

	result, err := jmespath.Search(queryParams, nodes)

	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(result)
	return bytes, err
}

func meshNodeToQueryNode(node mesh.MeshNode) *QueryNode {
	queryNode := new(QueryNode)
	queryNode.HostEndpoint = node.GetHostEndpoint()
	pubKey, _ := node.GetPublicKey()

	queryNode.PublicKey = pubKey.String()

	queryNode.WgEndpoint = node.GetWgEndpoint()
	queryNode.WgHost = node.GetWgHost().String()

	queryNode.Timestamp = node.GetTimeStamp()
	queryNode.Routes = node.GetRoutes()
	queryNode.Description = node.GetDescription()
	queryNode.Alias = node.GetAlias()

	return queryNode
}

func NewJmesQuerier(manager mesh.MeshManager) Querier {
	return &JmesQuerier{manager: manager}
}
