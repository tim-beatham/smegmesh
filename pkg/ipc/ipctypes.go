package ipc

import "github.com/tim-beatham/wgmesh/pkg/ctrlserver"

type JoinMeshArgs struct {
	MeshId   string
	IpAdress string
}

type GetMeshReply struct {
	Nodes []ctrlserver.MeshNode
}
