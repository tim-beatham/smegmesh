package robin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/query"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

type IpcHandler struct {
	Server ctrlserver.CtrlServer
}

func (n *IpcHandler) CreateMesh(args *ipc.NewMeshArgs, reply *string) error {
	meshId, err := n.Server.GetMeshManager().CreateMesh(args.WgPort)

	if err != nil {
		return err
	}

	err = n.Server.GetMeshManager().AddSelf(&mesh.AddSelfParams{
		MeshId:   meshId,
		WgPort:   args.WgPort,
		Endpoint: args.Endpoint,
	})

	if err != nil {
		return err
	}

	*reply = meshId
	return err
}

func (n *IpcHandler) ListMeshes(_ string, reply *ipc.ListMeshReply) error {
	meshNames := make([]string, len(n.Server.GetMeshManager().GetMeshes()))

	i := 0
	for meshId, _ := range n.Server.GetMeshManager().GetMeshes() {
		meshNames[i] = meshId
		i++
	}

	*reply = ipc.ListMeshReply{Meshes: meshNames}
	return nil
}

func (n *IpcHandler) JoinMesh(args ipc.JoinMeshArgs, reply *string) error {
	peerConnection, err := n.Server.GetConnectionManager().GetConnection(args.IpAdress)

	if err != nil {
		return err
	}

	client, err := peerConnection.GetClient()

	if err != nil {
		return err
	}

	c := rpc.NewMeshCtrlServerClient(client)

	if err != nil {
		return err
	}

	configuration := n.Server.GetConfiguration()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(configuration.Timeout))
	defer cancel()

	meshReply, err := c.GetMesh(ctx, &rpc.GetMeshRequest{MeshId: args.MeshId})

	if err != nil {
		return err
	}

	err = n.Server.GetMeshManager().AddMesh(&mesh.AddMeshParams{
		MeshId:    args.MeshId,
		WgPort:    args.Port,
		MeshBytes: meshReply.Mesh,
	})

	if err != nil {
		return err
	}

	err = n.Server.GetMeshManager().AddSelf(&mesh.AddSelfParams{
		MeshId:   args.MeshId,
		WgPort:   args.Port,
		Endpoint: args.Endpoint,
	})

	if err != nil {
		return err
	}

	*reply = strconv.FormatBool(true)
	return nil
}

// LeaveMesh leaves a mesh network
func (n *IpcHandler) LeaveMesh(meshId string, reply *string) error {
	err := n.Server.GetMeshManager().LeaveMesh(meshId)

	if err == nil {
		*reply = fmt.Sprintf("Left Mesh %s", meshId)
	}

	return err
}

func (n *IpcHandler) GetMesh(meshId string, reply *ipc.GetMeshReply) error {
	theMesh := n.Server.GetMeshManager().GetMesh(meshId)

	if theMesh == nil {
		return fmt.Errorf("mesh %s does not exist", meshId)
	}

	meshSnapshot, err := theMesh.GetMesh()

	if err != nil {
		return err
	}

	if theMesh == nil {
		return errors.New("mesh does not exist")
	}

	nodes := make([]ctrlserver.MeshNode, len(meshSnapshot.GetNodes()))

	i := 0
	for _, node := range meshSnapshot.GetNodes() {
		pubKey, _ := node.GetPublicKey()

		if err != nil {
			return err
		}

		node := ctrlserver.MeshNode{
			HostEndpoint: node.GetHostEndpoint(),
			WgEndpoint:   node.GetWgEndpoint(),
			PublicKey:    pubKey.String(),
			WgHost:       node.GetWgHost().String(),
			Timestamp:    node.GetTimeStamp(),
			Routes: lib.Map(node.GetRoutes(), func(r mesh.Route) ctrlserver.MeshRoute {
				return ctrlserver.MeshRoute{
					Destination: r.GetDestination().String(),
					Path:        r.GetPath(),
				}
			}),
			Description: node.GetDescription(),
			Alias:       node.GetAlias(),
			Services:    node.GetServices(),
		}

		nodes[i] = node
		i += 1
	}

	*reply = ipc.GetMeshReply{Nodes: nodes}
	return nil
}

func (n *IpcHandler) GetDOT(meshId string, reply *string) error {
	g := mesh.NewMeshDotConverter(n.Server.GetMeshManager())

	result, err := g.Generate(meshId)

	if err != nil {
		return err
	}

	*reply = result
	return nil
}

func (n *IpcHandler) Query(params ipc.QueryMesh, reply *string) error {
	queryResponse, err := n.Server.GetQuerier().Query(params.MeshId, params.Query)

	if err != nil {
		return err
	}

	*reply = string(queryResponse)
	return nil
}

func (n *IpcHandler) PutDescription(description string, reply *string) error {
	err := n.Server.GetMeshManager().SetDescription(description)

	if err != nil {
		return err
	}

	*reply = fmt.Sprintf("Set description to %s", description)
	return nil
}

func (n *IpcHandler) PutAlias(alias string, reply *string) error {
	err := n.Server.GetMeshManager().SetAlias(alias)

	if err != nil {
		return err
	}

	*reply = fmt.Sprintf("Set alias to %s", alias)
	return nil
}

func (n *IpcHandler) PutService(service ipc.PutServiceArgs, reply *string) error {
	err := n.Server.GetMeshManager().SetService(service.Service, service.Value)

	if err != nil {
		return err
	}

	*reply = "success"
	return nil
}

func (n *IpcHandler) DeleteService(service string, reply *string) error {
	err := n.Server.GetMeshManager().RemoveService(service)

	if err != nil {
		return err
	}

	*reply = "success"
	return nil
}

func (n *IpcHandler) GetNode(args ipc.GetNodeArgs, reply *string) error {
	node := n.Server.GetMeshManager().GetNode(args.MeshId, args.NodeId)

	if node == nil {
		*reply = "nil"
		return nil
	}

	queryNode := query.MeshNodeToQueryNode(node)

	bytes, err := json.Marshal(queryNode)

	if err != nil {
		*reply = err.Error()
		return nil
	}

	*reply = string(bytes)
	return nil
}

type RobinIpcParams struct {
	CtrlServer ctrlserver.CtrlServer
}

func NewRobinIpc(ipcParams RobinIpcParams) IpcHandler {
	return IpcHandler{
		Server: ipcParams.CtrlServer,
	}
}
