package main

import (
	"fmt"
	ipcRpc "net/rpc"
	"os"

	"github.com/akamensky/argparse"
	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
	graph "github.com/tim-beatham/smegmesh/pkg/dot"
	"github.com/tim-beatham/smegmesh/pkg/ipc"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

type CreateMeshParams struct {
	Client           *ipcRpc.Client
	Endpoint         string
	WgArgs           ipc.WireGuardArgs
	AdvertiseRoutes  bool
	AdvertiseDefault bool
}

func createMesh(client *ipc.SmegmeshIpc, args *ipc.NewMeshArgs) {
	var reply string
	err := client.CreateMesh(args, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func listMeshes(client *ipc.SmegmeshIpc) {
	reply := new(ipc.ListMeshReply)

	err := client.ListMeshes(reply)

	if err != nil {
		logging.Log.WriteErrorf(err.Error())
		return
	}

	for _, meshId := range reply.Meshes {
		fmt.Println(meshId)
	}
}

func joinMesh(client *ipc.SmegmeshIpc, args ipc.JoinMeshArgs) {
	var reply string

	err := client.JoinMesh(args, &reply)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(reply)
}

func leaveMesh(client *ipc.SmegmeshIpc, meshId string) {
	var reply string

	err := client.LeaveMesh(meshId, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func getGraph(client *ipc.SmegmeshIpc) {
	listMeshesReply := new(ipc.ListMeshReply)

	err := client.ListMeshes(listMeshesReply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	meshes := make(map[string][]ctrlserver.MeshNode)

	for _, meshId := range listMeshesReply.Meshes {
		var meshReply ipc.GetMeshReply

		err := client.GetMesh(meshId, &meshReply)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		meshes[meshId] = meshReply.Nodes
	}

	dotGenerator := graph.NewMeshGraphConverter(meshes)
	dot, err := dotGenerator.Generate()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(dot)
}

func queryMesh(client *ipc.SmegmeshIpc, meshId, query string) {
	var reply string

	args := ipc.QueryMesh{
		MeshId: meshId,
		Query:  query,
	}

	err := client.Query(args, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func putDescription(client *ipc.SmegmeshIpc, meshId, description string) {
	var reply string

	err := client.PutDescription(ipc.PutDescriptionArgs{
		MeshId:      meshId,
		Description: description,
	}, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

// putAlias: puts an alias for the node
func putAlias(client *ipc.SmegmeshIpc, meshid, alias string) {
	var reply string

	err := client.PutAlias(ipc.PutAliasArgs{
		MeshId: meshid,
		Alias:  alias,
	}, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func setService(client *ipc.SmegmeshIpc, meshId, service, value string) {
	var reply string

	err := client.PutService(ipc.PutServiceArgs{
		MeshId:  meshId,
		Service: service,
		Value:   value,
	}, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func deleteService(client *ipc.SmegmeshIpc, meshId, service string) {
	var reply string

	err := client.DeleteService(ipc.DeleteServiceArgs{
		MeshId:  meshId,
		Service: service,
	}, &reply)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(reply)
}

func main() {
	parser := argparse.NewParser("smgctl",
		"smegctl Manipulate WireGuard mesh networks")

	newMeshCmd := parser.NewCommand("new-mesh", "Create a new mesh")
	listMeshCmd := parser.NewCommand("list-meshes", "List meshes the node is connected to")
	joinMeshCmd := parser.NewCommand("join-mesh", "Join a mesh network")
	getGraphCmd := parser.NewCommand("get-graph", "Convert a mesh into DOT format")
	leaveMeshCmd := parser.NewCommand("leave-mesh", "Leave a mesh network")
	queryMeshCmd := parser.NewCommand("query-mesh", "Query a mesh network using JMESPath")
	putDescriptionCmd := parser.NewCommand("put-description", "Place a description for the node")
	putAliasCmd := parser.NewCommand("put-alias", "Place an alias for the node")
	setServiceCmd := parser.NewCommand("set-service", "Place a service into your advertisements")
	deleteServiceCmd := parser.NewCommand("delete-service", "Remove a service from your advertisements")

	var newMeshPort *int = newMeshCmd.Int("p", "wgport", &argparse.Options{
		Default: 0,
		Help:    "WireGuard port to use to the interface. A default of 0 uses an unused ephmeral port.",
	})

	var newMeshEndpoint *string = newMeshCmd.String("e", "endpoint", &argparse.Options{
		Help: "Publicly routeable endpoint to advertise within the mesh",
	})

	var newMeshRole *string = newMeshCmd.Selector("r", "role", []string{"peer", "client"}, &argparse.Options{
		Help: "Role in the mesh network. A value of peer means that the node is publicly routeable and thus considered" +
			" in the gossip protocol. Client means that the node is not publicly routeable and is not a candidate in the gossip" +
			" protocol",
	})
	var newMeshKeepAliveWg *int = newMeshCmd.Int("k", "KeepAliveWg", &argparse.Options{
		Default: 0,
		Help:    "WireGuard KeepAlive value for NAT traversal and firewall holepunching",
	})

	var newMeshAdvertiseRoutes *bool = newMeshCmd.Flag("a", "advertise", &argparse.Options{
		Help: "Advertise routes to other mesh network into the mesh",
	})

	var newMeshAdvertiseDefaults *bool = newMeshCmd.Flag("d", "defaults", &argparse.Options{
		Help: "Advertise ::/0 into the mesh network",
	})

	var joinMeshId *string = joinMeshCmd.String("m", "meshid", &argparse.Options{
		Required: true,
		Help:     "MeshID of the mesh network to join",
	})

	var joinMeshIpAddress *string = joinMeshCmd.String("i", "ip", &argparse.Options{
		Required: true,
		Help:     "IP address of the bootstrapping node to join through",
	})

	var joinMeshEndpoint *string = joinMeshCmd.String("e", "endpoint", &argparse.Options{
		Help: "Publicly routeable endpoint to advertise within the mesh",
	})

	var joinMeshRole *string = joinMeshCmd.Selector("r", "role", []string{"peer", "client"}, &argparse.Options{
		Help: "Role in the mesh network. A value of peer means that the node is publicly routeable and thus considered" +
			" in the gossip protocol. Client means that the node is not publicly routeable and is not a candidate in the gossip" +
			" protocol",
	})

	var joinMeshPort *int = joinMeshCmd.Int("p", "wgport", &argparse.Options{
		Default: 0,
		Help:    "WireGuard port to use to the interface. A default of 0 uses an unused ephmeral port.",
	})

	var joinMeshKeepAliveWg *int = joinMeshCmd.Int("k", "KeepAliveWg", &argparse.Options{
		Default: 0,
		Help:    "WireGuard KeepAlive value for NAT traversal and firewall ho;lepunching",
	})

	var joinMeshAdvertiseRoutes *bool = joinMeshCmd.Flag("a", "advertise", &argparse.Options{
		Help: "Advertise routes to other mesh network into the mesh",
	})

	var joinMeshAdvertiseDefaults *bool = joinMeshCmd.Flag("d", "defaults", &argparse.Options{
		Help: "Advertise ::/0 into the mesh network",
	})

	var leaveMeshMeshId *string = leaveMeshCmd.String("m", "mesh", &argparse.Options{
		Required: true,
		Help:     "MeshID of the mesh to leave",
	})

	var queryMeshMeshId *string = queryMeshCmd.String("m", "mesh", &argparse.Options{
		Required: true,
		Help:     "MeshID of the mesh to query",
	})
	var queryMeshQuery *string = queryMeshCmd.String("q", "query", &argparse.Options{
		Required: true,
		Help:     "JMESPath Query Of The Mesh Network To Query",
	})

	var description *string = putDescriptionCmd.String("d", "description", &argparse.Options{
		Required: true,
		Help:     "Description of the node in the mesh",
	})

	var descriptionMeshId *string = putDescriptionCmd.String("m", "meshid", &argparse.Options{
		Required: true,
		Help:     "MeshID of the mesh network to join",
	})

	var aliasMeshId *string = putAliasCmd.String("m", "meshid", &argparse.Options{
		Required: true,
		Help:     "MeshID of the mesh network to join",
	})

	var alias *string = putAliasCmd.String("a", "alias", &argparse.Options{
		Required: true,
		Help:     "Alias of the node to set can be used in DNS to lookup an IP address",
	})

	var serviceKey *string = setServiceCmd.String("s", "service", &argparse.Options{
		Required: true,
		Help:     "Key of the service to advertise in the mesh network",
	})
	var serviceValue *string = setServiceCmd.String("v", "value", &argparse.Options{
		Required: true,
		Help:     "Value of the service to advertise in the mesh network",
	})

	var serviceMeshId *string = setServiceCmd.String("m", "meshid", &argparse.Options{
		Required: true,
		Help:     "MeshID of the mesh network to join",
	})

	var deleteServiceKey *string = deleteServiceCmd.String("s", "service", &argparse.Options{
		Required: true,
		Help:     "Key of the service to remove",
	})

	var deleteServiceMeshid *string = deleteServiceCmd.String("m", "meshid", &argparse.Options{
		Required: true,
		Help:     "MeshID of the mesh network to join",
	})

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	client, err := ipc.NewClientIpc()
	if err != nil {
		panic(err)
	}

	if newMeshCmd.Happened() {
		args := &ipc.NewMeshArgs{
			WgArgs: ipc.WireGuardArgs{
				Endpoint:              *newMeshEndpoint,
				Role:                  *newMeshRole,
				WgPort:                *newMeshPort,
				KeepAliveWg:           *newMeshKeepAliveWg,
				AdvertiseDefaultRoute: *newMeshAdvertiseDefaults,
				AdvertiseRoutes:       *newMeshAdvertiseRoutes,
			},
		}

		createMesh(client, args)
	}

	if listMeshCmd.Happened() {
		listMeshes(client)
	}

	if joinMeshCmd.Happened() {
		args := ipc.JoinMeshArgs{
			IpAddress: *joinMeshIpAddress,
			MeshId:    *joinMeshId,
			WgArgs: ipc.WireGuardArgs{
				Endpoint:              *joinMeshEndpoint,
				Role:                  *joinMeshRole,
				WgPort:                *joinMeshPort,
				KeepAliveWg:           *joinMeshKeepAliveWg,
				AdvertiseDefaultRoute: *joinMeshAdvertiseDefaults,
				AdvertiseRoutes:       *joinMeshAdvertiseRoutes,
			},
		}
		joinMesh(client, args)
	}

	if getGraphCmd.Happened() {
		getGraph(client)
	}

	if leaveMeshCmd.Happened() {
		leaveMesh(client, *leaveMeshMeshId)
	}

	if queryMeshCmd.Happened() {
		queryMesh(client, *queryMeshMeshId, *queryMeshQuery)
	}

	if putDescriptionCmd.Happened() {
		putDescription(client, *descriptionMeshId, *description)
	}

	if putAliasCmd.Happened() {
		putAlias(client, *aliasMeshId, *alias)
	}

	if setServiceCmd.Happened() {
		setService(client, *serviceMeshId, *serviceKey, *serviceValue)
	}

	if deleteServiceCmd.Happened() {
		deleteService(client, *deleteServiceMeshid, *deleteServiceKey)
	}
}
