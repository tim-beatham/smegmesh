package ctrlserver

import (
	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/conn"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"github.com/tim-beatham/smegmesh/pkg/query"
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

func (c *CtrlServerStub) GetConfiguration() *conf.DaemonConfiguration {
	return &conf.DaemonConfiguration{
		GrpcPort:          8080,
		BaseConfiguration: conf.WgConfiguration{},
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
