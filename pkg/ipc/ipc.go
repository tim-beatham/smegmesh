package ipc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
)

// WireGuardArgs are provided args specific to WireGuard
type WireGuardArgs struct {
	// WgPort is the WireGuard port to expose
	WgPort int
	// KeepAliveWg is the number of seconds to keep alive
	// for WireGuard NAT/firewall traversal
	KeepAliveWg int
	// AdvertiseRoutes whether or not to advertise routes to and from the
	// mesh network
	AdvertiseRoutes bool
	// AdvertiseDefaultRoute whether or not to advertise the default route
	// into the mesh network
	AdvertiseDefaultRoute bool
	// Endpoint is the routable alias of the machine. Can be an IP
	// or DNS entry
	Endpoint string
	// Role is the role of the individual in the mesh
	Role string
}

type NewMeshArgs struct {
	// WgArgs are specific WireGuard args to use
	WgArgs WireGuardArgs
}

type JoinMeshArgs struct {
	// MeshId is the ID of the mesh to join
	MeshId string
	// IpAddress is a routable IP in another mesh
	IpAdress string
	// WgArgs is the WireGuard parameters to use.
	WgArgs WireGuardArgs
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

type MeshIpc interface {
	CreateMesh(args *NewMeshArgs, reply *string) error
	ListMeshes(name string, reply *ListMeshReply) error
	JoinMesh(args JoinMeshArgs, reply *string) error
	LeaveMesh(meshId string, reply *string) error
	GetMesh(meshId string, reply *GetMeshReply) error
	Query(query QueryMesh, reply *string) error
	PutDescription(description string, reply *string) error
	PutAlias(alias string, reply *string) error
	PutService(args PutServiceArgs, reply *string) error
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
