package ipc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	ipcRpc "net/rpc"
	"os"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
)

const SockAddr = "/tmp/wgmesh_sock"

type MeshIpc interface {
	CreateMesh(args *NewMeshArgs, reply *string) error
	ListMeshes(name string, reply *ListMeshReply) error
	JoinMesh(args *JoinMeshArgs, reply *string) error
	LeaveMesh(meshId string, reply *string) error
	GetMesh(meshId string, reply *GetMeshReply) error
	Query(query QueryMesh, reply *string) error
	PutDescription(args PutDescriptionArgs, reply *string) error
	PutAlias(args PutAliasArgs, reply *string) error
	PutService(args PutServiceArgs, reply *string) error
	DeleteService(args DeleteServiceArgs, reply *string) error
}

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
	IpAddress string
	// WgArgs is the WireGuard parameters to use.
	WgArgs WireGuardArgs
}

type PutServiceArgs struct {
	Service string
	Value   string
	MeshId  string
}

type DeleteServiceArgs struct {
	Service string
	MeshId  string
}

type PutAliasArgs struct {
	Alias  string
	MeshId string
}

type PutDescriptionArgs struct {
	Description string
	MeshId      string
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

type ClientIpc struct {
	client *ipcRpc.Client
}

func NewClientIpc() (*ClientIpc, error) {
	client, err := ipcRpc.DialHTTP("unix", SockAddr)

	if err != nil {
		return nil, err
	}

	return &ClientIpc{
		client: client,
	}, nil
}

func (c *ClientIpc) CreateMesh(args *NewMeshArgs, reply *string) error {
	return c.client.Call("IpcHandler.CreateMesh", args, reply)
}

func (c *ClientIpc) ListMeshes(reply *ListMeshReply) error {
	return c.client.Call("IpcHandler.ListMeshes", "", reply)
}

func (c *ClientIpc) JoinMesh(args JoinMeshArgs, reply *string) error {
	return c.client.Call("IpcHandler.JoinMesh", &args, reply)
}

func (c *ClientIpc) LeaveMesh(meshId string, reply *string) error {
	return c.client.Call("IpcHandler.LeaveMesh", &meshId, reply)
}

func (c *ClientIpc) GetMesh(meshId string, reply *GetMeshReply) error {
	return c.client.Call("IpcHandler.GetMesh", &meshId, reply)
}

func (c *ClientIpc) Query(query QueryMesh, reply *string) error {
	return c.client.Call("IpcHandler.Query", &query, reply)
}

func (c *ClientIpc) PutDescription(args PutDescriptionArgs, reply *string) error {
	return c.client.Call("IpcHandler.PutDescription", &args, reply)
}

func (c *ClientIpc) PutAlias(args PutAliasArgs, reply *string) error {
	return c.client.Call("IpcHandler.PutAlias", &args, reply)
}

func (c *ClientIpc) PutService(args PutServiceArgs, reply *string) error {
	return c.client.Call("IpcHandler.PutService", &args, reply)
}

func (c *ClientIpc) DeleteService(args DeleteServiceArgs, reply *string) error {
	return c.client.Call("IpcHandler.DeleteService", &args, reply)
}

func (c *ClientIpc) Close() error {
	return c.Close()
}

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
