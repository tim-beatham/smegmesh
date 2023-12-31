package robin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
)

type IpcHandler struct {
	Server ctrlserver.CtrlServer
}

func getOverrideConfiguration(args *ipc.WireGuardArgs) conf.WgConfiguration {
	overrideConf := conf.WgConfiguration{}

	if args.Role != "" {
		role := conf.NodeType(args.Role)
		overrideConf.Role = &role
	}

	if args.Endpoint != "" {
		overrideConf.Endpoint = &args.Endpoint
	}

	if args.KeepAliveWg != 0 {
		keepAliveWg := args.KeepAliveWg
		overrideConf.KeepAliveWg = &keepAliveWg
	}

	overrideConf.AdvertiseRoutes = &args.AdvertiseRoutes
	overrideConf.AdvertiseDefaultRoute = &args.AdvertiseDefaultRoute
	return overrideConf
}

func (n *IpcHandler) CreateMesh(args *ipc.NewMeshArgs, reply *string) error {
	overrideConf := getOverrideConfiguration(&args.WgArgs)

	meshId, err := n.Server.GetMeshManager().CreateMesh(&mesh.CreateMeshParams{
		Port: args.WgArgs.WgPort,
		Conf: &overrideConf,
	})

	if err != nil {
		return err
	}

	err = n.Server.GetMeshManager().AddSelf(&mesh.AddSelfParams{
		MeshId:   meshId,
		WgPort:   args.WgArgs.WgPort,
		Endpoint: args.WgArgs.Endpoint,
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
	for meshId := range n.Server.GetMeshManager().GetMeshes() {
		meshNames[i] = meshId
		i++
	}

	*reply = ipc.ListMeshReply{Meshes: meshNames}
	return nil
}

func (n *IpcHandler) JoinMesh(args *ipc.JoinMeshArgs, reply *string) error {
	overrideConf := getOverrideConfiguration(&args.WgArgs)

	if n.Server.GetMeshManager().GetMesh(args.MeshId) != nil {
		return fmt.Errorf("user is already apart of the mesh")
	}

	peerConnection, err := n.Server.GetConnectionManager().GetConnection(args.IpAddress)

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
		WgPort:    args.WgArgs.WgPort,
		MeshBytes: meshReply.Mesh,
		Conf:      &overrideConf,
	})

	if err != nil {
		return err
	}

	err = n.Server.GetMeshManager().AddSelf(&mesh.AddSelfParams{
		MeshId:   args.MeshId,
		WgPort:   args.WgArgs.WgPort,
		Endpoint: args.WgArgs.Endpoint,
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
		node := ctrlserver.NewCtrlNode(theMesh, node)

		nodes[i] = *node
		i += 1
	}

	*reply = ipc.GetMeshReply{Nodes: nodes}
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

func (n *IpcHandler) PutDescription(args ipc.PutDescriptionArgs, reply *string) error {
	err := n.Server.GetMeshManager().SetDescription(args.MeshId, args.Description)

	if err != nil {
		return err
	}

	*reply = fmt.Sprintf("set description to %s for %s", args.Description, args.MeshId)
	return nil
}

func (n *IpcHandler) PutAlias(args ipc.PutAliasArgs, reply *string) error {
	if args.Alias == "" {
		return fmt.Errorf("alias not provided")
	}

	err := n.Server.GetMeshManager().SetAlias(args.MeshId, args.Alias)

	if err != nil {
		return err
	}

	*reply = fmt.Sprintf("Set alias to %s", args.Alias)
	return nil
}

func (n *IpcHandler) PutService(service ipc.PutServiceArgs, reply *string) error {
	err := n.Server.GetMeshManager().SetService(service.MeshId, service.Service, service.Value)

	if err != nil {
		return err
	}

	*reply = "success"
	return nil
}

func (n *IpcHandler) DeleteService(service ipc.DeleteServiceArgs, reply *string) error {
	err := n.Server.GetMeshManager().RemoveService(service.MeshId, service.Service)

	if err != nil {
		return err
	}

	*reply = "success"
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
