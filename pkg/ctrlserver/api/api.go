package api

import (
	"github.com/gin-gonic/gin"
	ctrlserver "github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver/api/mesh"
)

func RunAPI(server *ctrlserver.MeshCtrlServer) *gin.Engine {
	r := gin.Default()

	r.POST("/mesh", func(ctx *gin.Context) {
		mesh.JoinMesh(ctx, server)
	})

	return r
}
