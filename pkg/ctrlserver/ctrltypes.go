package ctrlserver

import "github.com/gin-gonic/gin"

/*
 * Represents a WireGuard node
 */
type MeshNode struct {
	Host      string
	CtrlPort  string
	WgPort    string
	WgHost    string
	GinServer gin.Engine
}

/*
 * Defines the mesh control server this node
 * is running
 */
type MeshCtrlServer struct {
	Host   string
	Port   int
	Meshes map[string]MeshNode
}
