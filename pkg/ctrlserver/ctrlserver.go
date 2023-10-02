/*
 * ctrlserver controls the WireGuard mesh. Contains an IpcHandler for
 * handling commands fired by wgmesh command.
 * Contains an RpcHandler for handling commands fired by another server.
 */
package ctrlserver

import (
	"errors"
	"net"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/conn"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/rpc"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type NewCtrlServerParams struct {
	WgClient     *wgctrl.Client
	Conf         *conf.WgMeshConfiguration
	AuthProvider rpc.AuthenticationServer
	CtrlProvider rpc.MeshCtrlServerServer
}

/*
 * NewCtrlServer creates a new instance of the ctrlserver.
 * It is associated with a WireGuard client and an interface.
 * wgClient: Represents the WireGuard control client.
 * ifName: WireGuard interface name
 */
func NewCtrlServer(params *NewCtrlServerParams) (*MeshCtrlServer, error) {
	ctrlServer := new(MeshCtrlServer)
	ctrlServer.Meshes = make(map[string]Mesh)
	ctrlServer.Client = params.WgClient
	ctrlServer.IfName = params.Conf.IfName

	connManagerParams := conn.NewConnectionManagerParams{
		CertificatePath:      params.Conf.CertificatePath,
		PrivateKey:           params.Conf.PrivateKeyPath,
		SkipCertVerification: params.Conf.SkipCertVerification,
	}

	connMgr, err := conn.NewConnectionManager(&connManagerParams)

	if err != nil {
		return nil, err
	}

	ctrlServer.ConnectionManager = connMgr

	connServerParams := conn.NewConnectionServerParams{
		CertificatePath:      params.Conf.CertificatePath,
		PrivateKey:           params.Conf.PrivateKeyPath,
		SkipCertVerification: params.Conf.SkipCertVerification,
		AuthProvider:         params.AuthProvider,
		CtrlProvider:         params.CtrlProvider,
	}

	connServer, err := conn.NewConnectionServer(&connServerParams)

	if err != nil {
		return nil, err
	}

	ctrlServer.ConnectionServer = connServer
	return ctrlServer, nil
}

/*
 * MeshExists returns true if the client is part of the mesh
 * false otherwise.
 */
func (server *MeshCtrlServer) MeshExists(meshId string) bool {
	_, inMesh := server.Meshes[meshId]
	return inMesh
}

/*
 * CreateMesh creates a new mesh instance, adds it to the map
 * of meshes and returns the newly created mesh.
 */
func (server *MeshCtrlServer) CreateMesh() (*Mesh, error) {
	key, err := wgtypes.GenerateKey()

	if err != nil {
		return nil, err
	}

	var mesh Mesh = Mesh{
		SharedKey: &key,
		Nodes:     make(map[string]MeshNode),
	}

	server.Meshes[key.String()] = mesh
	return &mesh, nil
}

/*
 * AddHostArgs parameters needed to add
 * a host to the mesh network
 */
type AddHostArgs struct {
	// HostEndpoint: Public IP and port of the gRPC server the node is running
	HostEndpoint string
	// PubilcKey: WireGuard public key of the device
	PublicKey string
	// MeshId: Mesh ID of the node the the host is joining
	MeshId string
	// WgEndpoint: Public IP and port of the WireGuard server that is running
	WgEndpoint string
	// WgIp: SLAAC generated WireGuard IP of the node
	WgIp string
}

/*
 * AddHost adds a host to the mesh
 */
func (server *MeshCtrlServer) AddHost(args AddHostArgs) error {
	nodes, contains := server.Meshes[args.MeshId]

	if !contains {
		return errors.New("The mesh: " + args.MeshId + " does not exist")
	}

	_, contains = nodes.Nodes[args.HostEndpoint]

	if contains {
		return errors.New("The node already has an endpoint in the mesh network")
	}

	node := MeshNode{
		HostEndpoint: args.HostEndpoint,
		WgEndpoint:   args.WgEndpoint,
		PublicKey:    args.PublicKey,
		WgHost:       args.WgIp,
	}

	err := server.AddWgPeer(node)

	if err == nil {
		nodes.Nodes[args.HostEndpoint] = node
	}

	return err
}

/*
 * GetDevice gets the WireGuard client associated with the
 * interface name.
 */
func (server *MeshCtrlServer) GetDevice() *wgtypes.Device {
	dev, err := server.Client.Device(server.IfName)

	if err != nil {
		return nil
	}

	return dev
}

/*
 * AddWgPeer Updates the WireGuard configuration to include the peer
 */
func (server *MeshCtrlServer) AddWgPeer(node MeshNode) error {
	peer := make([]wgtypes.PeerConfig, 1)

	peerPublic, err := wgtypes.ParseKey(node.PublicKey)

	if err != nil {
		return err
	}

	peerEndpoint, err := net.ResolveUDPAddr("udp", node.WgEndpoint)

	if err != nil {
		return err
	}

	allowedIps := make([]net.IPNet, 1)
	_, ipnet, err := net.ParseCIDR(node.WgHost)

	if err != nil {
		return err
	}

	allowedIps[0] = *ipnet

	peer[0] = wgtypes.PeerConfig{
		PublicKey:  peerPublic,
		Endpoint:   peerEndpoint,
		AllowedIPs: allowedIps,
	}

	cfg := wgtypes.Config{
		Peers: peer,
	}

	server.Client.ConfigureDevice(server.IfName, cfg)

	if err != nil {
		return err
	}

	return nil
}

/*
 * EnableInterface: Enables the given WireGuard interface.
 */
func (s *MeshCtrlServer) EnableInterface(meshId string) error {
	mesh, contains := s.Meshes[meshId]

	if !contains {
		return errors.New("Mesh does not exist")
	}

	endPoint := lib.GetOutboundIP().String() + ":8080"

	node, contains := mesh.Nodes[endPoint]

	if !contains {
		return errors.New("Node does not exist in the mesh")
	}

	return wg.EnableInterface(s.IfName, node.WgHost)
}
