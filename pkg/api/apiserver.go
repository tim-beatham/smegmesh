package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
	"github.com/tim-beatham/smegmesh/pkg/ipc"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/what8words"
)

type ApiServer interface {
	GetMeshes(c *gin.Context)
	Run(addr string) error
}

type SmegServer struct {
	router *gin.Engine
	client *ipc.SmegmeshIpc
	words  *what8words.What8Words
}

func (s *SmegServer) routeToApiRoute(meshNode ctrlserver.MeshNode) []Route {
	routes := make([]Route, len(meshNode.Routes))

	for index, route := range meshNode.Routes {

		if route.Path == nil {
			route.Path = make([]string, 0)
		}

		routes[index] = Route{
			Prefix: route.Destination,
			Path:   route.Path,
		}
	}

	return routes
}

func (s *SmegServer) meshNodeToAPIMeshNode(meshNode ctrlserver.MeshNode) *SmegNode {
	if meshNode.Routes == nil {
		meshNode.Routes = make([]ctrlserver.MeshRoute, 0)
	}

	alias := meshNode.Alias

	if alias == "" {
		alias, _ = s.words.ConvertIdentifier(meshNode.WgHost)
	}

	return &SmegNode{
		WgHost:      meshNode.WgHost,
		WgEndpoint:  meshNode.WgEndpoint,
		Endpoint:    meshNode.HostEndpoint,
		Timestamp:   int(meshNode.Timestamp),
		Description: meshNode.Description,
		Routes:      s.routeToApiRoute(meshNode),
		PublicKey:   meshNode.PublicKey,
		Alias:       alias,
		Services:    meshNode.Services,
		Stats: SmegStats{
			TotalTransmit:     meshNode.Stats.TransmitBytes,
			TotalReceived:     meshNode.Stats.ReceivedBytes,
			KeepAliveInterval: meshNode.Stats.PersistentKeepAliveInterval,
			AllowedIps:        meshNode.Stats.AllowedIPs,
		},
	}
}

func (s *SmegServer) meshToAPIMesh(meshId string, nodes []ctrlserver.MeshNode) SmegMesh {
	var smegMesh SmegMesh
	smegMesh.MeshId = meshId
	smegMesh.Nodes = make(map[string]SmegNode)

	for _, node := range nodes {
		smegMesh.Nodes[node.WgHost] = *s.meshNodeToAPIMeshNode(node)
	}

	return smegMesh
}

// CreateMesh: creates a mesh network
func (s *SmegServer) CreateMesh(c *gin.Context) {
	var createMesh CreateMeshRequest

	if err := c.ShouldBindJSON(&createMesh); err != nil {
		c.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})

		return
	}

	ipcRequest := ipc.NewMeshArgs{
		WgArgs: ipc.WireGuardArgs{
			WgPort: createMesh.WgPort,
		},
	}

	var reply string

	err := s.client.CreateMesh(&ipcRequest, &reply)

	if err != nil {
		c.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"meshid": reply,
	})
}

// JoinMesh: joins a mesh network
func (s *SmegServer) JoinMesh(c *gin.Context) {
	var joinMesh JoinMeshRequest

	if err := c.ShouldBindJSON(&joinMesh); err != nil {
		c.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})
		return
	}

	ipcRequest := ipc.JoinMeshArgs{
		MeshId:    joinMesh.MeshId,
		IpAddress: joinMesh.Bootstrap,
		WgArgs: ipc.WireGuardArgs{
			WgPort: joinMesh.WgPort,
		},
	}

	var reply string

	err := s.client.JoinMesh(ipcRequest, &reply)

	if err != nil {
		c.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &gin.H{
		"status": "success",
	})
}

// GetMesh: given a meshId returns the corresponding mesh
// network.
func (s *SmegServer) GetMesh(c *gin.Context) {
	meshidParam := c.Param("meshid")

	var meshid string = meshidParam

	getMeshReply := new(ipc.GetMeshReply)

	err := s.client.GetMesh(meshid, getMeshReply)

	if err != nil {
		c.JSON(http.StatusNotFound,
			&gin.H{
				"error": fmt.Sprintf("could not find mesh %s", meshidParam),
			})
		return
	}

	mesh := s.meshToAPIMesh(meshidParam, getMeshReply.Nodes)

	c.JSON(http.StatusOK, mesh)
}

func (s *SmegServer) GetMeshes(c *gin.Context) {
	listMeshesReply := new(ipc.ListMeshReply)

	err := s.client.ListMeshes(listMeshesReply)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	meshes := make([]SmegMesh, 0)

	for _, mesh := range listMeshesReply.Meshes {
		getMeshReply := new(ipc.GetMeshReply)

		err := s.client.GetMesh(mesh, getMeshReply)

		if err != nil {
			logging.Log.WriteErrorf(err.Error())
			c.JSON(http.StatusBadRequest, nil)
			return
		}

		meshes = append(meshes, s.meshToAPIMesh(mesh, getMeshReply.Nodes))
	}

	c.JSON(http.StatusOK, meshes)
}

func (s *SmegServer) Run(addr string) error {
	logging.Log.WriteInfof("Running API server")
	return s.router.Run(addr)
}

func NewSmegServer(conf ApiServerConf) (ApiServer, error) {
	client, err := ipc.NewClientIpc()

	if err != nil {
		return nil, err
	}

	words, err := what8words.NewWhat8Words(conf.WordsFile)

	if err != nil {
		return nil, err
	}

	router := gin.Default()

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: logging.Log.Writer(),
	}))

	smegServer := &SmegServer{
		router: router,
		client: client,
		words:  words,
	}

	router.GET("/meshes", smegServer.GetMeshes)
	router.GET("/mesh/:meshid", smegServer.GetMesh)
	router.POST("/mesh/create", smegServer.CreateMesh)
	router.POST("/mesh/join", smegServer.JoinMesh)
	return smegServer, nil
}
