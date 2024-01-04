package ctrlserver

import (
	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/conn"
	"github.com/tim-beatham/smegmesh/pkg/crdt"
	"github.com/tim-beatham/smegmesh/pkg/ip"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	logging "github.com/tim-beatham/smegmesh/pkg/log"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"github.com/tim-beatham/smegmesh/pkg/query"
	"github.com/tim-beatham/smegmesh/pkg/rpc"
	"github.com/tim-beatham/smegmesh/pkg/sync"
	"github.com/tim-beatham/smegmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
)

// NewCtrlServerParams are the params requried to create a new ctrl server
type NewCtrlServerParams struct {
	Conf         *conf.DaemonConfiguration
	Client       *wgctrl.Client
	CtrlProvider rpc.MeshCtrlServerServer
	SyncProvider rpc.SyncServiceServer
	Querier      query.Querier
}

// Create a new instance of the MeshCtrlServer or error if the
// operation failed
func NewCtrlServer(params *NewCtrlServerParams) (*MeshCtrlServer, error) {
	ctrlServer := new(MeshCtrlServer)
	meshFactory := &crdt.TwoPhaseMapFactory{
		Config: params.Conf,
	}
	nodeFactory := &crdt.MeshNodeFactory{
		Config: *params.Conf,
	}
	idGenerator := &lib.ShortIDGenerator{}
	ipAllocator := &ip.ULABuilder{}
	interfaceManipulator := wg.NewWgInterfaceManipulator(params.Client)

	ctrlServer.timers = make([]*lib.Timer, 0)

	configApplyer := mesh.NewWgMeshConfigApplyer()

	var syncer sync.Syncer

	meshManagerParams := &mesh.NewMeshManagerParams{
		Conf:                 *params.Conf,
		Client:               params.Client,
		MeshProvider:         meshFactory,
		NodeFactory:          nodeFactory,
		IdGenerator:          idGenerator,
		IPAllocator:          ipAllocator,
		InterfaceManipulator: interfaceManipulator,
		ConfigApplyer:        configApplyer,
		OnDelete: func(mesh mesh.MeshProvider) {
			_, err := syncer.Sync(mesh)

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
			}
		},
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

	syncer = sync.NewSyncer(&sync.NewSyncerParams{
		MeshManager:       ctrlServer.MeshManager,
		ConnectionManager: ctrlServer.ConnectionManager,
		Configuration:     params.Conf,
	})

	// Check any syncs every 1 second
	syncTimer := lib.NewTimer(func() error {
		err = syncer.SyncMeshes()

		if err != nil {
			logging.Log.WriteErrorf(err.Error())
		}

		return nil
	}, 1)

	heartbeatTimer := lib.NewTimer(func() error {
		logging.Log.WriteInfof("checking heartbeat")
		return ctrlServer.MeshManager.UpdateTimeStamp()
	}, params.Conf.HeartBeat)

	ctrlServer.timers = append(ctrlServer.timers, syncTimer, heartbeatTimer)

	ctrlServer.Querier = query.NewJmesQuerier(ctrlServer.MeshManager)
	ctrlServer.ConnectionServer = connServer

	for _, timer := range ctrlServer.timers {
		go timer.Run()
	}

	return ctrlServer, nil
}

func (s *MeshCtrlServer) GetConfiguration() *conf.DaemonConfiguration {
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

	for _, timer := range s.timers {
		err := timer.Stop()

		if err != nil {
			logging.Log.WriteErrorf(err.Error())
		}
	}

	return nil
}
