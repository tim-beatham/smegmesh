package ctrlserver

import (
	crdt "github.com/tim-beatham/wgmesh/pkg/automerge"
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
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

	configApplyer := mesh.NewWgMeshConfigApplyer()

	meshManagerParams := &mesh.NewMeshManagerParams{
		Conf:                 *params.Conf,
		Client:               params.Client,
		MeshProvider:         &meshFactory,
		NodeFactory:          &nodeFactory,
		IdGenerator:          idGenerator,
		IPAllocator:          ipAllocator,
		InterfaceManipulator: interfaceManipulator,
		ConfigApplyer:        configApplyer,
	}

	ctrlServer.MeshManager = mesh.NewMeshManager(meshManagerParams)
	configApplyer.SetMeshManager(ctrlServer.MeshManager)

	ctrlServer.Conf = params.Conf
	connManagerParams := conn.NewConnectionManagerParams{
		CertificatePath:      params.Conf.CertificatePath,
		PrivateKey:           params.Conf.PrivateKeyPath,
		SkipCertVerification: params.Conf.SkipCertVerification,
		CaCert:               params.Conf.CaCertificatePath,
		ConnFactory:          conn.NewWgCtrlConnection,
	}

	connMgr, err := conn.NewConnectionManager(&connManagerParams)

	if err != nil {
		return nil, err
	}

	ctrlServer.ConnectionManager = connMgr
	connServerParams := conn.NewConnectionServerParams{
		Conf:         params.Conf,
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

func (s *MeshCtrlServer) GetConfiguration() *conf.WgMeshConfiguration {
	return s.Conf
}

func (s *MeshCtrlServer) GetClient() *wgctrl.Client {
	return s.Client
}

func (s *MeshCtrlServer) GetQuerier() query.Querier {
	return s.Querier
}

func (s *MeshCtrlServer) GetMeshManager() mesh.MeshManager {
	return s.MeshManager
}

func (s *MeshCtrlServer) GetConnectionManager() conn.ConnectionManager {
	return s.ConnectionManager
}

// Close closes the ctrl server tearing down any connections that exist
func (s *MeshCtrlServer) Close() error {
	if err := s.ConnectionManager.Close(); err != nil {
		logging.Log.WriteErrorf(err.Error())
	}

	if err := s.MeshManager.Close(); err != nil {
		logging.Log.WriteErrorf(err.Error())
	}

	if err := s.ConnectionServer.Close(); err != nil {
		logging.Log.WriteErrorf(err.Error())
	}

	return nil
}
