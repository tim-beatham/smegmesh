package robin

import (
	"testing"

	"github.com/tim-beatham/wgmesh/pkg/ctrlserver"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
)

func getRequester() *IpcHandler {
	return &IpcHandler{Server: ctrlserver.NewCtrlServerStub()}
}

func TestCreateMeshRepliesMeshId(t *testing.T) {
	var reply string
	requester := getRequester()

	err := requester.CreateMesh(&ipc.NewMeshArgs{
		IfName:   "wg0",
		WgPort:   5000,
		Endpoint: "abc.com",
	}, &reply)

	if err != nil {
		t.Error(err)
	}

	if len(reply) == 0 {
		t.Fatalf(`reply should have been returned`)
	}
}

func TestListMeshesNoMeshesListsEmpty(t *testing.T) {
	var reply ipc.ListMeshReply
	requester := getRequester()

	err := requester.ListMeshes("", &reply)

	if err != nil {
		t.Error(err)
	}

	if len(reply.Meshes) != 0 {
		t.Fatalf(`meshes should be empty`)
	}
}

func TestListMeshesMeshesNotEmpty(t *testing.T) {
	var reply ipc.ListMeshReply
	requester := getRequester()

	requester.Server.GetMeshManager().AddMesh(&mesh.AddMeshParams{
		MeshId:    "tim123",
		DevName:   "wg0",
		WgPort:    5000,
		MeshBytes: make([]byte, 0),
	})

	err := requester.ListMeshes("", &reply)

	if err != nil {
		t.Error(err)
	}

	if len(reply.Meshes) != 1 {
		t.Fatalf(`only only mesh exists`)
	}

	if reply.Meshes[0] != "tim123" {
		t.Fatalf(`meshId was %s expected %s`, reply.Meshes[0], "tim123")
	}
}
