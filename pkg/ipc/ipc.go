package ipc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
)

type NewMeshArgs struct {
	// WgPort is the WireGuard port to expose
	WgPort int
	// Endpoint is the routable alias of the machine. Can be an IP
	// or DNS entry
	Endpoint string
	Role     string
}

type JoinMeshArgs struct {
	// MeshId is the ID of the mesh to join
	MeshId string
	// IpAddress is a routable IP in another mesh
	IpAdress string
	// Port is the WireGuard port to expose
	Port int
	// Endpoint to use to override the default
	Endpoint string
	// Client specifies whether we should join as a client of the peer
	// we are connecting to
	Client bool
	Role   string
}

type PutServiceArgs struct {
	Service string
	Value   string
}

type GetMeshReply struct {
	Nodes []ctrlserver.MeshNode
}

type ListMeshReply struct {
	Meshes []string
}

type QueryMesh struct {
	MeshId string
	Query  string
}

type GetNodeArgs struct {
	NodeId string
	MeshId string
}

type MeshIpc interface {
	CreateMesh(args *NewMeshArgs, reply *string) error
	ListMeshes(name string, reply *ListMeshReply) error
	JoinMesh(args JoinMeshArgs, reply *string) error
	LeaveMesh(meshId string, reply *string) error
	GetMesh(meshId string, reply *GetMeshReply) error
	GetDOT(meshId string, reply *string) error
	Query(query QueryMesh, reply *string) error
	PutDescription(description string, reply *string) error
	PutAlias(alias string, reply *string) error
	PutService(args PutServiceArgs, reply *string) error
	GetNode(args GetNodeArgs, reply *string) error
	DeleteService(service string, reply *string) error
}

const SockAddr = "/tmp/wgmesh_ipc.sock"

func RunIpcHandler(server MeshIpc) error {
	if err := os.RemoveAll(SockAddr); err != nil {
		return errors.New("could not find to address")
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
