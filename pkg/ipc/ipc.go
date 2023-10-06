package ipc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"

	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
)

type JoinMeshArgs struct {
	MeshId   string
	IpAdress string
}

type GetMeshReply struct {
	Nodes []crdt.MeshNodeCrdt
}

type ListMeshReply struct {
	Meshes []string
}

type MeshIpc interface {
	CreateMesh(name string, reply *string) error
	ListMeshes(name string, reply *ListMeshReply) error
	JoinMesh(args JoinMeshArgs, reply *string) error
	GetMesh(meshId string, reply *GetMeshReply) error
	EnableInterface(meshId string, reply *string) error
}

const SockAddr = "/tmp/wgmesh_ipc.sock"

func RunIpcHandler(server MeshIpc) error {
	if err := os.RemoveAll(SockAddr); err != nil {
		return errors.New("Could not find to address")
	}

	rpc.Register(server)
	rpc.HandleHTTP()

	l, e := net.Listen("unix", SockAddr)
	if e != nil {
		return e
	}

	http.Serve(l, nil)
	return nil
}