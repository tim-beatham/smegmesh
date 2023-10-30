package ctrlserver

import (
	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"github.com/tim-beatham/wgmesh/pkg/query"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
)

// NewCtrlServerParams are the params requried to create a new ctrl server
type NewCtrlServerParams struct {
	Conf         *conf.WgMeshConfiguration
	Client       *wgctrl.Client
	AuthProvider rpc.AuthenticationServer
	CtrlProvider rpc.MeshCtrlServerServer
	SyncProvider rpc.SyncServiceServer
	Querier      query.Querier
}

// Create a new instance of the MeshCtrlServer or error if the
// operation failed
func NewCtrlServer(params *NewCtrlServerParams) (*MeshCtrlServer, error) {
	ctrlServer := new(MeshCtrlServer)
	meshFactory := crdt.CrdtProviderFactory{}
	nodeFactory := crdt.MeshNodeFactory{
		Config: *params.Conf,
	}
	idGenerator := &lib.UUIDGenerator{}
	ipAllocator := &ip.ULABuilder{}
	interfaceManipulator := wg.NewWgInterfaceManipulator(params.Client)

	meshManagerParams := &mesh.NewMeshManagerParams{
		Conf:                 *params.Conf,
		Client:               params.Client,
		MeshProvider:         &meshFactory,
		NodeFactory:          &nodeFactory,
		IdGenerator:          idGenerator,
		IPAllocator:          ipAllocator,
		InterfaceManipulator: interfaceManipulator,
	}

	ctrlServer.MeshManager = mesh.NewMeshManager(meshManagerParams)

	ctrlServer.Conf = params.Conf
	connManagerParams := conn.NewConnectionManageParams{
		CertificatePath:      params.Conf.CertificatePath,
		PrivateKey:           params.Conf.PrivateKeyPath,
		SkipCertVerification: params.Conf.SkipCertVerification,
		CaCert:               params.Conf.CaCertificatePath,
	}

	connMgr, err := conn.NewConnectionManager(&connManagerParams)

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

	ctrlServer.Querier = query.NewJmesQuerier(ctrlServer.MeshManager)
	ctrlServer.ConnectionServer = connServer

	return ctrlServer, nil
}

// Close closes the ctrl server tearing down any connections that exist
func (s *MeshCtrlServer) Close() error {
	if err := s.ConnectionManager.Close(); err != nil {
		return err
	}

	if err := s.ConnectionServer.Close(); err != nil {
		return err
	}

	return nil
}
