package main

import (
	"fmt"
	ipcRpc "net/rpc"
	"os"

	"github.com/akamensky/argparse"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

type CreateMeshParams struct {
	Client   *ipcRpc.Client
	WgPort   int
	Endpoint string
	Role     string
}

func createMesh(args *CreateMeshParams) string {
	var reply string
	newMeshParams := ipc.NewMeshArgs{
		WgPort:   args.WgPort,
		Endpoint: args.Endpoint,
		Role:     args.Role,
	}

	err := args.Client.Call("IpcHandler.CreateMesh", &newMeshParams, &reply)

	if err != nil {
		return err.Error()
	}

	return reply
}

func listMeshes(client *ipcRpc.Client) {
	reply := new(ipc.ListMeshReply)

	err := client.Call("IpcHandler.ListMeshes", "", &reply)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return
	}

	for _, meshId := range reply.Meshes {
		fmt.Println(meshId)
	}
}

type JoinMeshParams struct {
	Client    *ipcRpc.Client
	MeshId    string
	IpAddress string
	IfName    string
	WgPort    int
	Endpoint  string
	Role      string
}

func joinMesh(params *JoinMeshParams) string {
	var reply string

	args := ipc.JoinMeshArgs{
		MeshId:   params.MeshId,
		IpAdress: params.IpAddress,
		Port:     params.WgPort,
		Role:     params.Role,
	}

	err := params.Client.Call("IpcHandler.JoinMesh", &args, &reply)

	if err != nil {
		return err.Error()
	}

	return reply
}

func leaveMesh(client *ipcRpc.Client, meshId string) {
	var reply string

	err := client.Call("IpcHandler.LeaveMesh", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func enableInterface(client *ipcRpc.Client, meshId string) {
	var reply string

	err := client.Call("IpcHandler.EnableInterface", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func getGraph(client *ipcRpc.Client, meshId string) {
	var reply string

	err := client.Call("IpcHandler.GetDOT", &meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func queryMesh(client *ipcRpc.Client, meshId, query string) {
	var reply string

	err := client.Call("IpcHandler.Query", &ipc.QueryMesh{MeshId: meshId, Query: query}, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

// putDescription: puts updates the description about the node to the meshes
func putDescription(client *ipcRpc.Client, description string) {
	var reply string

	err := client.Call("IpcHandler.PutDescription", &description, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

// putAlias: puts an alias for the node
func putAlias(client *ipcRpc.Client, alias string) {
	var reply string

	err := client.Call("IpcHandler.PutAlias", &alias, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func setService(client *ipcRpc.Client, service, value string) {
	var reply string

	serviceArgs := &ipc.PutServiceArgs{
		Service: service,
		Value:   value,
	}

	err := client.Call("IpcHandler.PutService", serviceArgs, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func deleteService(client *ipcRpc.Client, service string) {
	var reply string

	err := client.Call("IpcHandler.PutService", &service, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func getNode(client *ipcRpc.Client, nodeId, meshId string) {
	var reply string
	args := &ipc.GetNodeArgs{
		NodeId: nodeId,
		MeshId: meshId,
	}

	err := client.Call("IpcHandler.GetNode", &args, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func main() {
	parser := argparse.NewParser("wg-mesh",
		"wg-mesh Manipulate WireGuard meshes")

	newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
	listMeshCmd := parser.NewCommand("list-meshes", "List meshes the node is connected to")
	joinMeshCmd := parser.NewCommand("join-mesh", "Join a mesh network")
	enableInterfaceCmd := parser.NewCommand("enable-interface", "Enable A Specific Mesh Interface")
	getGraphCmd := parser.NewCommand("get-graph", "Convert a mesh into DOT format")
	leaveMeshCmd := parser.NewCommand("leave-mesh", "Leave a mesh network")
	queryMeshCmd := parser.NewCommand("query-mesh", "Query a mesh network using JMESPath")
	putDescriptionCmd := parser.NewCommand("put-description", "Place a description for the node")
	putAliasCmd := parser.NewCommand("put-alias", "Place an alias for the node")
	setServiceCmd := parser.NewCommand("set-service", "Place a service into your advertisements")
	deleteServiceCmd := parser.NewCommand("delete-service", "Remove a service from your advertisements")
	getNodeCmd := parser.NewCommand("get-node", "Get a specific node from the mesh")

	var newMeshPort *int = newMeshCmd.Int("p", "wgport", &argparse.Options{})
	var newMeshEndpoint *string = newMeshCmd.String("e", "endpoint", &argparse.Options{})
	var newMeshRole *string = newMeshCmd.Selector("r", "role", []string{"peer", "client"}, &argparse.Options{})

	var joinMeshId *string = joinMeshCmd.String("m", "mesh", &argparse.Options{Required: true})
	var joinMeshIpAddress *string = joinMeshCmd.String("i", "ip", &argparse.Options{Required: true})
	var joinMeshPort *int = joinMeshCmd.Int("p", "wgport", &argparse.Options{})
	var joinMeshEndpoint *string = joinMeshCmd.String("e", "endpoint", &argparse.Options{})
	var joinMeshRole *string = joinMeshCmd.Selector("r", "role", []string{"peer", "client"}, &argparse.Options{})

	var enableInterfaceMeshId *string = enableInterfaceCmd.String("m", "mesh", &argparse.Options{Required: true})

	var getGraphMeshId *string = getGraphCmd.String("m", "mesh", &argparse.Options{Required: true})

	var leaveMeshMeshId *string = leaveMeshCmd.String("m", "mesh", &argparse.Options{Required: true})

	var queryMeshMeshId *string = queryMeshCmd.String("m", "mesh", &argparse.Options{Required: true})
	var queryMeshQuery *string = queryMeshCmd.String("q", "query", &argparse.Options{Required: true})

	var description *string = putDescriptionCmd.String("d", "description", &argparse.Options{Required: true})

	var alias *string = putAliasCmd.String("a", "alias", &argparse.Options{Required: true})

	var serviceKey *string = setServiceCmd.String("s", "service", &argparse.Options{Required: true})
	var serviceValue *string = setServiceCmd.String("v", "value", &argparse.Options{Required: true})

	var deleteServiceKey *string = deleteServiceCmd.String("s", "service", &argparse.Options{Required: true})

	var getNodeNodeId *string = getNodeCmd.String("n", "nodeid", &argparse.Options{Required: true})
	var getNodeMeshId *string = getNodeCmd.String("m", "meshid", &argparse.Options{Required: true})

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	client, err := ipcRpc.DialHTTP("unix", SockAddr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if newMeshCmd.Happened() {
		fmt.Println(createMesh(&CreateMeshParams{
			Client:   client,
			WgPort:   *newMeshPort,
			Endpoint: *newMeshEndpoint,
			Role:     *newMeshRole,
		}))
	}

	if listMeshCmd.Happened() {
		listMeshes(client)
	}

	if joinMeshCmd.Happened() {
		fmt.Println(joinMesh(&JoinMeshParams{
			Client:    client,
			WgPort:    *joinMeshPort,
			IpAddress: *joinMeshIpAddress,
			MeshId:    *joinMeshId,
			Endpoint:  *joinMeshEndpoint,
			Role:      *joinMeshRole,
		}))
	}

	if getGraphCmd.Happened() {
		getGraph(client, *getGraphMeshId)
	}

	if enableInterfaceCmd.Happened() {
		enableInterface(client, *enableInterfaceMeshId)
	}

	if leaveMeshCmd.Happened() {
		leaveMesh(client, *leaveMeshMeshId)
	}

	if queryMeshCmd.Happened() {
		queryMesh(client, *queryMeshMeshId, *queryMeshQuery)
	}

	if putDescriptionCmd.Happened() {
		putDescription(client, *description)
	}

	if putAliasCmd.Happened() {
		putAlias(client, *alias)
	}

	if setServiceCmd.Happened() {
		setService(client, *serviceKey, *serviceValue)
	}

	if deleteServiceCmd.Happened() {
		deleteService(client, *deleteServiceKey)
	}

	if getNodeCmd.Happened() {
		getNode(client, *getNodeNodeId, *getNodeMeshId)
	}
}
