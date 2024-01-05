package ipc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	ipcRPC "net/rpc"
	"os"

	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
)

const SockAddr = "/tmp/smeg.sock"

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

// PutServiceArgs: args to place a service into the data store
type PutServiceArgs struct {
	Service string
	Value   string
	MeshId  string
}

// DeleteServiceArgs: args to remove a service from the data store
type DeleteServiceArgs struct {
	Service string
	MeshId  string
}

// PutAliasArgs: args to assign an alias to a node
type PutAliasArgs struct {
	// Alias: represents the alias of the node
	Alias string
	// MeshId: represents the meshID of the node
	MeshId string
}

// PutDescriptionArgs: args to assign a description to a node
type PutDescriptionArgs struct {
	// Description: descriptio to add to the network
	Description string
	// MeshID to add to the mesh network
	MeshId string
}

// GetMeshReply: ipc reply to get the mesh network
type GetMeshReply struct {
	Nodes []ctrlserver.MeshNode
}

// ListMeshReply: ipc reply of the networks the node is part of
type ListMeshReply struct {
	Meshes []string
}

// Querymesh: ipc args to query a mesh network
type QueryMesh struct {
	// MeshId: id of the mesh to query
	MeshId string
	// JMESPath: query string to query
	Query string
}

// ClientIpc: Framework to invoke ipc calls to the daemon
type ClientIpc interface {
	// CreateMesh: create a mesh network, return an error if the operation failed
	CreateMesh(args *NewMeshArgs, reply *string) error
	// ListMesh: list mesh network the node is a part of, return an error if the operation failed
	ListMeshes(args *ListMeshReply, reply *string) error
	// JoinMesh: join a mesh network return an error if the operation failed
	JoinMesh(args JoinMeshArgs, reply *string) error
	// LeaveMesh: leave a mesh network, return an error if the operation failed
	LeaveMesh(meshId string, reply *string) error
	// GetMesh: get the given mesh network, return an error if the operation failed
	GetMesh(meshId string, reply *GetMeshReply) error
	// Query: query the given mesh network
	Query(query QueryMesh, reply *string) error
	// PutDescription: assign a description to yourself
	PutDescription(args PutDescriptionArgs, reply *string) error
	// PutAlias: assign an alias to yourself
	PutAlias(args PutAliasArgs, reply *string) error
	// PutService: assign a service to yourself
	PutService(args PutServiceArgs, reply *string) error
	// DeleteService: retract a service
	DeleteService(args DeleteServiceArgs, reply *string) error
}

type SmegmeshIpc struct {
	client *ipcRPC.Client
}

func NewClientIpc() (*SmegmeshIpc, error) {
	client, err := ipcRPC.DialHTTP("unix", SockAddr)

	if err != nil {
		return nil, err
	}

	return &SmegmeshIpc{
		client: client,
	}, nil
}

func (c *SmegmeshIpc) CreateMesh(args *NewMeshArgs, reply *string) error {
	return c.client.Call("IpcHandler.CreateMesh", args, reply)
}

func (c *SmegmeshIpc) ListMeshes(reply *ListMeshReply) error {
	return c.client.Call("IpcHandler.ListMeshes", "", reply)
}

func (c *SmegmeshIpc) JoinMesh(args JoinMeshArgs, reply *string) error {
	return c.client.Call("IpcHandler.JoinMesh", &args, reply)
}

func (c *SmegmeshIpc) LeaveMesh(meshId string, reply *string) error {
	return c.client.Call("IpcHandler.LeaveMesh", &meshId, reply)
}

func (c *SmegmeshIpc) GetMesh(meshId string, reply *GetMeshReply) error {
	return c.client.Call("IpcHandler.GetMesh", &meshId, reply)
}

func (c *SmegmeshIpc) Query(query QueryMesh, reply *string) error {
	return c.client.Call("IpcHandler.Query", &query, reply)
}

func (c *SmegmeshIpc) PutDescription(args PutDescriptionArgs, reply *string) error {
	return c.client.Call("IpcHandler.PutDescription", &args, reply)
}

func (c *SmegmeshIpc) PutAlias(args PutAliasArgs, reply *string) error {
	return c.client.Call("IpcHandler.PutAlias", &args, reply)
}

func (c *SmegmeshIpc) PutService(args PutServiceArgs, reply *string) error {
	return c.client.Call("IpcHandler.PutService", &args, reply)
}

func (c *SmegmeshIpc) DeleteService(args DeleteServiceArgs, reply *string) error {
	return c.client.Call("IpcHandler.DeleteService", &args, reply)
}

func (c *SmegmeshIpc) Close() error {
	return c.client.Close()
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
