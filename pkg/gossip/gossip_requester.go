package gossip

import (
	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
)

type GossipRequester struct {
	Server      *ctrlserver.MeshCtrlServer
	ipAlloactor ip.IPAllocator
}

func (r *GossipRequester) CreateMesh(name string, reply *string) error {
	return nil
}

func (r *GossipRequester) ListMeshes(name string, reply string) error {
	return nil
}

func (r *GossipRequester) JoinMesh(args ipc.JoinMeshArgs, reply *string) error {
	return nil
}

func (r *GossipRequester) GetMesh(meshId string, reply *ipc.GetMeshReply) error {
	return nil
}
