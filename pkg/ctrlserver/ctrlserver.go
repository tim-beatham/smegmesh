/*
 * ctrlserver controls the WireGuard mesh. Contains an IpcHandler for
 * handling commands fired by wgmesh command.
 * Contains an RpcHandler for handling commands fired by another server.
 */
package ctrlserver

import (
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/manager"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"golang.zx2c4.com/wireguard/wgctrl"
)

type NewCtrlServerParams struct {
	WgClient     *wgctrl.Client
	Conf         *conf.WgMeshConfiguration
	AuthProvider rpc.AuthenticationServer
	CtrlProvider rpc.MeshCtrlServerServer
	SyncProvider rpc.SyncServiceServer
}

/*
 * NewCtrlServer creates a new instance of the ctrlserver.
 * It is associated with a WireGuard client and an interface.
 * wgClient: Represents the WireGuard control client.
 * ifName: WireGuard interface name
 */
func NewCtrlServer(params *NewCtrlServerParams) (*MeshCtrlServer, error) {
	ctrlServer := new(MeshCtrlServer)
	ctrlServer.Client = params.WgClient
	ctrlServer.MeshManager = manager.NewMeshManager(*params.WgClient)
	ctrlServer.Conf = params.Conf

	connManagerParams := conn.NewJwtConnectionManagerParams{
		CertificatePath:      params.Conf.CertificatePath,
		PrivateKey:           params.Conf.PrivateKeyPath,
		SkipCertVerification: params.Conf.SkipCertVerification,
	}

	connMgr, err := conn.NewJwtConnectionManager(&connManagerParams)

	if err != nil {
		return nil, err
	}

	ctrlServer.ConnectionManager = connMgr

	connServerParams := conn.NewConnectionServerParams{
		Conf:         params.Conf,
		AuthProvider: params.AuthProvider,
		CtrlProvider: params.CtrlProvider,
		SyncProvider: params.SyncProvider,
	}

	connServer, err := conn.NewConnectionServer(&connServerParams)

	if err != nil {
		return nil, err
	}

	ctrlServer.ConnectionServer = connServer
	return ctrlServer, nil
}
