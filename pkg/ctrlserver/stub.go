package ctrlserver

import (
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/query"
	"golang.zx2c4.com/wireguard/wgctrl"
)

type CtrlServerStub struct {
	manager           mesh.MeshManager
	querier           query.Querier
	connectionManager conn.ConnectionManager
}

func NewCtrlServerStub() *CtrlServerStub {
	var manager mesh.MeshManager = mesh.NewMeshManagerStub()
	return &CtrlServerStub{
		manager:           manager,
		querier:           query.NewJmesQuerier(manager),
		connectionManager: &conn.ConnectionManagerStub{},
	}
}

func (c *CtrlServerStub) GetConfiguration() *conf.WgMeshConfiguration {
	return &conf.WgMeshConfiguration{
		GrpcPort: "8080",
		Endpoint: "abc.com",
	}
}

func (c *CtrlServerStub) GetClient() *wgctrl.Client {
	return &wgctrl.Client{}
}

func (c *CtrlServerStub) GetQuerier() query.Querier {
	return c.querier
}

func (c *CtrlServerStub) GetMeshManager() mesh.MeshManager {
	return c.manager
}

func (c *CtrlServerStub) Close() error {
	return nil
}

func (c *CtrlServerStub) GetConnectionManager() conn.ConnectionManager {
	return c.connectionManager
}
