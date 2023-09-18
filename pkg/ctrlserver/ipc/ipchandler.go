package ipc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

type Mesh struct {
}

/*
 * Create a new WireGuard mesh network
 */
func (n Mesh) CreateNewMesh(name *string, reply *string) error {
	key, err := wgtypes.GenerateKey()

	if err != nil {
		return err
	}

	*reply = key.String()
	return nil
}

func RunIpcHandler() error {
	if err := os.RemoveAll(SockAddr); err != nil {
		return errors.New("Could not find to address")
	}

	newMeshIpc := new(Mesh)
	rpc.Register(newMeshIpc)
	rpc.HandleHTTP()

	l, e := net.Listen("unix", SockAddr)
	if e != nil {
		return e
	}

	http.Serve(l, nil)
	return nil
}
