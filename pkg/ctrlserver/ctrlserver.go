package ctrlserver

import (
	"errors"
	"net"

	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

/*
 * Create a new control server instance running
 * on the provided port.
 */
func NewCtrlServer(wgClient *wgctrl.Client, ifName string) *MeshCtrlServer {
	ctrlServer := new(MeshCtrlServer)
	ctrlServer.Meshes = make(map[string]Mesh)
	ctrlServer.Client = wgClient
	ctrlServer.IfName = ifName
	return ctrlServer
}

/*
 * Given the meshid returns true if the node is in the mesh
 * false otherwise.
 */
func (server *MeshCtrlServer) IsInMesh(meshId string) bool {
	_, inMesh := server.Meshes[meshId]
	return inMesh
}

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

type AddHostArgs struct {
	HostEndpoint string
	PublicKey    string
	MeshId       string
	WgEndpoint   string
	WgIp         string
}

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

	err := AddWgPeer(server.IfName, server.Client, node)

	if err == nil {
		nodes.Nodes[args.HostEndpoint] = node
	}

	return err
}

func (server *MeshCtrlServer) GetDevice() *wgtypes.Device {
	dev, err := server.Client.Device(server.IfName)

	if err != nil {
		return nil
	}

	return dev
}

func AddWgPeer(ifName string, client *wgctrl.Client, node MeshNode) error {
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

	err = client.ConfigureDevice(ifName, cfg)

	if err != nil {
		return err
	}

	return nil
}

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
