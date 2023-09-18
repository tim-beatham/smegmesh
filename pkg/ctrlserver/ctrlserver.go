package ctrlserver

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

/*
 * Create a new control server instance running
 * on the provided port.
 */
func NewCtrlServer(host string, port int) *MeshCtrlServer {
	ctrlServer := new(MeshCtrlServer)
	ctrlServer.Port = port
	ctrlServer.Meshes = make(map[string]MeshNode)
	ctrlServer.Host = host
	return ctrlServer
}

/*
 * Given the meshid returns true if the node is in the mesh
 * false otherwise.
 */
func (server *MeshCtrlServer) IsInMesh(meshId string) bool {
	_, inMesh := server.Meshes[meshId]
	return inMesh
}

func (server *MeshCtrlServer) GetEndpoint() string {
	return server.Host + ":" + strconv.Itoa(server.Port)
}

/*
 * Run the gin server instance
 */
func (server *MeshCtrlServer) Run() bool {
	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello",
		})
	})

	r.Run(server.GetEndpoint())
	return true
}
