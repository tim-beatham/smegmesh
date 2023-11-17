package api

import (
	"fmt"
	"net/http"

	ipcRpc "net/rpc"

	"github.com/gin-gonic/gin"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

type ApiServer interface {
	GetMeshes(c *gin.Context)
	Run(addr string) error
}

type SmegServer struct {
	router *gin.Engine
	client *ipcRpc.Client
}

func meshNodeToAPIMeshNode(meshNode ctrlserver.MeshNode) *SmegNode {
	if meshNode.Routes == nil {
		meshNode.Routes = make([]string, 0)
	}

	return &SmegNode{
		WgHost:      meshNode.WgHost,
		WgEndpoint:  meshNode.WgEndpoint,
		Endpoint:    meshNode.HostEndpoint,
		Timestamp:   int(meshNode.Timestamp),
		Description: meshNode.Description,
		Routes:      meshNode.Routes,
		PublicKey:   meshNode.PublicKey,
		Alias:       meshNode.Alias,
		Services:    meshNode.Services,
	}
}

func meshToAPIMesh(meshId string, nodes []ctrlserver.MeshNode) SmegMesh {
	var smegMesh SmegMesh
	smegMesh.MeshId = meshId
	smegMesh.Nodes = make(map[string]SmegNode)

	for _, node := range nodes {
		smegMesh.Nodes[node.WgHost] = *meshNodeToAPIMeshNode(node)
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
		IfName: createMesh.IfName,
		WgPort: createMesh.WgPort,
	}

	var reply string

	err := s.client.Call("IpcHandler.CreateMesh", &ipcRequest, &reply)

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
		MeshId:   joinMesh.MeshId,
		IpAdress: joinMesh.Bootstrap,
		IfName:   joinMesh.IfName,
		Port:     joinMesh.WgPort,
	}

	var reply string

	err := s.client.Call("IpcHandler.JoinMesh", &ipcRequest, &reply)

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

	err := s.client.Call("IpcHandler.GetMesh", &meshid, &getMeshReply)

	if err != nil {
		c.JSON(http.StatusNotFound,
			&gin.H{
				"error": fmt.Sprintf("could not find mesh %s", meshidParam),
			})
		return
	}

	mesh := meshToAPIMesh(meshidParam, getMeshReply.Nodes)

	c.JSON(http.StatusOK, mesh)
}

func (s *SmegServer) GetMeshes(c *gin.Context) {
	listMeshesReply := new(ipc.ListMeshReply)

	err := s.client.Call("IpcHandler.ListMeshes", "", &listMeshesReply)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	meshes := make([]SmegMesh, 0)

	for _, mesh := range listMeshesReply.Meshes {
		getMeshReply := new(ipc.GetMeshReply)

		err := s.client.Call("IpcHandler.GetMesh", &mesh, &getMeshReply)

		if err != nil {
			logging.Log.WriteErrorf(err.Error())
			c.JSON(http.StatusBadRequest, nil)
			return
		}

		meshes = append(meshes, meshToAPIMesh(mesh, getMeshReply.Nodes))
	}

	c.JSON(http.StatusOK, meshes)
}

func (s *SmegServer) Run(addr string) error {
	logging.Log.WriteInfof("Running API server")
	return s.router.Run(addr)
}

func NewSmegServer() (ApiServer, error) {
	client, err := ipcRpc.DialHTTP("unix", SockAddr)

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
	}

	router.GET("/meshes", smegServer.GetMeshes)
	router.GET("/mesh/:meshid", smegServer.GetMesh)
	router.POST("/mesh/create", smegServer.CreateMesh)
	router.POST("/mesh/join", smegServer.JoinMesh)
	return smegServer, nil
}
