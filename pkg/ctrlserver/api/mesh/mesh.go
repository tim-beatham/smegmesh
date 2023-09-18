package mesh

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
)

type JoinMeshInput struct {
	MeshId string `json:"mesh-id" binding:"required`
}

func JoinMesh(c *gin.Context, server *ctrlserver.MeshCtrlServer) {
	var input JoinMeshInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
